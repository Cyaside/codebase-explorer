package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

const compactCombinedSizeThreshold = 27000
const projectedOutputBytesPerFunction = 1500
const workerEvidenceLimit = 40000

type batchOutcome struct {
	results      []fullai.FunctionResult
	err          error
	promptBytes  int
	promptTokens int
	outputTokens int
}

func (s Service) executeAdaptiveAI(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, functions fullai.Functions, evidence fullai.Evidence, summary fullai.Summary) (fullai.Execution, fullai.Summary, error) {
	started := time.Now()
	config := s.providerConfig(request)
	client, err := s.providers.ClientFor(config)
	if err != nil {
		return fullai.Execution{}, summary, fmt.Errorf("connection: %w", err)
	}
	jobs := make([]fullai.FunctionJob, 0, len(functions.Jobs))
	var dashboard fullai.FunctionJob
	for _, job := range functions.Jobs {
		if job.Name == "dashboard" {
			dashboard = job
		} else {
			jobs = append(jobs, job)
		}
	}
	if len(jobs) == 0 {
		return fullai.Execution{}, summary, fmt.Errorf("no AI functions were prepared")
	}
	readableEvidence := 0
	for _, item := range evidence.Items {
		if item.ReadStatus == "read" {
			readableEvidence++
		}
	}
	if readableEvidence == 0 {
		return fullai.Execution{}, summary, fmt.Errorf("AI analysis failed: no readable repository evidence was selected")
	}
	packHash, err := fullai.PackHash()
	if err != nil {
		return fullai.Execution{}, summary, err
	}
	cacheKey, keyErr := cache.BuildAIExecutionKey(config, evidence, packHash, s.settings.AppVersion)
	if s.cache.Enabled() && keyErr == nil {
		cached, loadErr := s.cache.LoadAIExecution(cacheKey)
		if loadErr == nil && cached.Execution.Status == "executed" && cached.Execution.PackHash == packHash {
			execution := cached.Execution
			execution.GeneratedAt = analysis.GeneratedAt
			execution.CallCount = 0
			execution.RepairCalls = 0
			execution.PromptBytes = 0
			execution.DurationMS = time.Since(started).Milliseconds()
			execution.Note = "reused verified AI execution from cache"
			summary.Status = execution.Status
			summary.ExecutedFunctions = execution.ExecutedCount
			summary.VerifiedFunctions = execution.VerifiedCount
			summary = attachExecutionMetrics(summary, execution)
			summary.CacheStatus = cache.StatusHit
			emitAnalyzeProgress(request, "provider-cache", cache.StatusHit, "reused verified AI execution")
			return execution, summary, nil
		}
	}
	summary.CacheStatus = cache.StatusMiss
	groups := groupAIJobs(jobs, evidence)
	initialCalls := len(groups)
	mode := "one call"
	if initialCalls == 2 {
		mode = "two parallel calls"
	}
	emitAnalyzeProgress(request, "full-ai-execution", "running", fmt.Sprintf("running %s for %d functions", mode, len(jobs)))
	execution := fullai.Execution{
		SchemaVersion: fullai.ExecutionSchemaVersion, GeneratedAt: analysis.GeneratedAt,
		PackHash: packHash,
		Mode:     string(fullai.ModeFull), Provider: config.Name, Model: config.Model,
		Status: "running", Results: make([]fullai.FunctionResult, 0, len(functions.Jobs)),
	}
	outcomes := make(chan batchOutcome, len(groups))
	for _, group := range groups {
		group := group
		go func() { outcomes <- runAIBatch(ctx, client, config, group, evidence, workerEvidenceLimit) }()
	}
	byName := make(map[string]fullai.FunctionResult, len(jobs))
	for range groups {
		outcome := <-outcomes
		execution.PromptBytes += outcome.promptBytes
		execution.PromptTokens += outcome.promptTokens
		execution.OutputTokens += outcome.outputTokens
		if errors.Is(outcome.err, context.Canceled) || ctx.Err() != nil {
			return fullai.Execution{}, summary, ctx.Err()
		}
		for _, result := range outcome.results {
			byName[result.Name] = result
		}
	}
	failed := make([]fullai.FunctionJob, 0)
	for _, job := range jobs {
		if byName[job.Name].Status != "verified" {
			failed = append(failed, job)
		}
	}
	if len(failed) > 0 {
		emitAnalyzeProgress(request, "full-ai-retry", "running", fmt.Sprintf("retrying %d failed or unverified section(s) once", len(failed)))
		outcome := runAIBatch(ctx, client, config, failed, evidence, 12000)
		execution.PromptBytes += outcome.promptBytes
		execution.PromptTokens += outcome.promptTokens
		execution.OutputTokens += outcome.outputTokens
		execution.RepairCalls = 1
		if errors.Is(outcome.err, context.Canceled) || ctx.Err() != nil {
			return fullai.Execution{}, summary, ctx.Err()
		}
		for _, result := range outcome.results {
			if result.Status == "verified" || byName[result.Name].Status == "" {
				byName[result.Name] = result
			}
		}
	}
	verifiedWorkers := 0
	for _, job := range jobs {
		result := byName[job.Name]
		if result.Status == "verified" {
			verifiedWorkers++
		}
		execution.Results = append(execution.Results, result)
		emitAnalyzeProgress(request, "full-ai-function", result.Status, fmt.Sprintf("%s: %s", job.Name, fullAIResultDetail(result)))
	}
	if verifiedWorkers == 0 {
		emitAnalyzeProgress(request, "full-ai-execution", "failed", "no AI section passed verification")
		return fullai.Execution{}, summary, fmt.Errorf("AI analysis failed: no section passed evidence and structure verification")
	}
	if dashboard.Name != "" {
		result := assembleDashboard(dashboard, execution.Results)
		execution.Results = append(execution.Results, result)
	}
	execution.ExecutedCount, execution.VerifiedCount, execution.FailedCount = countFullAIResults(execution.Results)
	execution.CallCount = initialCalls + execution.RepairCalls
	execution.DurationMS = time.Since(started).Milliseconds()
	execution.Status = fullAIExecutionStatus(execution)
	execution.Note = fmt.Sprintf("%d provider call(s), including %d repair; %d section(s) passed checks, %d incomplete", execution.CallCount, execution.RepairCalls, execution.VerifiedCount, len(execution.Results)-execution.VerifiedCount)
	summary.Status = execution.Status
	summary.ExecutedFunctions = execution.ExecutedCount
	summary.VerifiedFunctions = execution.VerifiedCount
	summary = attachExecutionMetrics(summary, execution)
	summary.Note = appendFullAINote(summary.Note, execution.Note)
	if execution.Status == "executed" && s.cache.Enabled() && keyErr == nil {
		_ = s.cache.SaveAIExecution(cache.AIExecutionPayload{CachedAt: time.Now().UTC(), Key: cacheKey, Execution: execution})
	}
	emitAnalyzeProgress(request, "full-ai-execution", execution.Status, execution.Note)
	return execution, summary, nil
}

func attachExecutionMetrics(summary fullai.Summary, execution fullai.Execution) fullai.Summary {
	summary.PackVersion = fullai.PackVersion
	summary.PackHash = execution.PackHash
	summary.Provider = execution.Provider
	summary.Model = execution.Model
	summary.CallCount = execution.CallCount
	summary.RepairCalls = execution.RepairCalls
	summary.PromptBytes = execution.PromptBytes
	summary.PromptTokens = execution.PromptTokens
	summary.OutputTokens = execution.OutputTokens
	summary.DurationMS = execution.DurationMS
	return summary
}

func groupAIJobs(jobs []fullai.FunctionJob, evidence fullai.Evidence) [][]fullai.FunctionJob {
	bytes := 0
	for _, item := range evidence.Items {
		bytes += len(item.Snippet)
	}
	if bytes+len(jobs)*projectedOutputBytesPerFunction <= compactCombinedSizeThreshold {
		return [][]fullai.FunctionJob{jobs}
	}
	groups := make([][]fullai.FunctionJob, 2)
	for _, job := range jobs {
		if job.Name == "flowchart" || job.Name == "issues" {
			groups[1] = append(groups[1], job)
		} else {
			groups[0] = append(groups[0], job)
		}
	}
	if len(groups[0]) == 0 || len(groups[1]) == 0 {
		return [][]fullai.FunctionJob{jobs}
	}
	return groups
}

func runAIBatch(ctx context.Context, client provider.Client, config provider.Config, jobs []fullai.FunctionJob, evidence fullai.Evidence, evidenceLimit int) batchOutcome {
	prompt, err := fullai.BuildBatchPrompt(jobs, evidence, evidenceLimit)
	if err != nil {
		return failedBatch(jobs, err, config.APIKey)
	}
	response, err := client.Complete(ctx, provider.PromptRequest{Config: config, SystemPrompt: prompt.SystemPrompt, UserPrompt: prompt.UserPrompt})
	if err != nil {
		failed := failedBatch(jobs, err, config.APIKey)
		failed.promptBytes = len(prompt.SystemPrompt) + len(prompt.UserPrompt)
		return failed
	}
	outputs, err := fullai.ParseBatchOutput(fullai.RedactSensitiveText(stripAPIKey(response.Content, config.APIKey)))
	if err != nil {
		failed := failedBatch(jobs, err, config.APIKey)
		failed.promptBytes = len(prompt.SystemPrompt) + len(prompt.UserPrompt)
		return failed
	}
	results := make([]fullai.FunctionResult, 0, len(jobs))
	for _, job := range jobs {
		result := fullai.FunctionResult{Name: job.Name, InstructionPath: job.InstructionPath, EvidencePaths: append([]string(nil), job.EvidencePaths...)}
		instructionHash, loadErr := fullai.InstructionHash(job)
		if loadErr == nil {
			result.InstructionHash = instructionHash
		}
		output, ok := outputs[job.Name]
		if !ok {
			result.Status = "failed"
			result.Error = "provider omitted or malformed this function output"
		} else {
			output = fullai.RedactFunctionOutput(output)
			output, report := fullai.VerifyFunctionOutputDetailed(job, output)
			result.Output = output
			result.Verification = report
			result.Verified = report.Verified
			result.Status = report.Status
		}
		results = append(results, result)
	}
	return batchOutcome{results: results, promptBytes: len(prompt.SystemPrompt) + len(prompt.UserPrompt), promptTokens: response.PromptTokens, outputTokens: response.OutputTokens}
}

func failedBatch(jobs []fullai.FunctionJob, err error, apiKey string) batchOutcome {
	results := make([]fullai.FunctionResult, 0, len(jobs))
	for _, job := range jobs {
		results = append(results, fullai.FunctionResult{
			Name: job.Name, Status: "failed", InstructionPath: job.InstructionPath,
			EvidencePaths: append([]string(nil), job.EvidencePaths...), Error: fullai.RedactSensitiveText(stripAPIKey(err.Error(), apiKey)),
		})
	}
	return batchOutcome{results: results, err: err}
}

func stripAPIKey(value, apiKey string) string {
	if apiKey == "" {
		return value
	}
	return strings.ReplaceAll(value, apiKey, "[REDACTED API KEY]")
}

func assembleDashboard(job fullai.FunctionJob, results []fullai.FunctionResult) fullai.FunctionResult {
	output := fullai.FunctionOutput{}
	for _, result := range results {
		if !result.Verified {
			continue
		}
		if output.Summary == "" && strings.TrimSpace(result.Output.Summary) != "" {
			output.Summary = result.Output.Summary
		}
		output.KeyFindings = append(output.KeyFindings, result.Output.KeyFindings...)
		output.Recommendations = append(output.Recommendations, result.Output.Recommendations...)
		output.GraphEdges = append(output.GraphEdges, result.Output.GraphEdges...)
		output.IssueSignals = append(output.IssueSignals, result.Output.IssueSignals...)
	}
	output, verification := fullai.VerifyFunctionOutputDetailed(job, output)
	status := verification.Status
	return fullai.FunctionResult{Name: job.Name, Status: status, InstructionPath: job.InstructionPath,
		EvidencePaths: job.EvidencePaths, Output: output, Verification: verification, Verified: verification.Verified}
}
