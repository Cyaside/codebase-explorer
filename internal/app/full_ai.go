package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func (s Service) buildFullAIPlan(request AnalyzeRequest, scanResult repo.ScanResult, analysis analyzer.Result, changeResult changes.Result, supportFiles []string) (fullai.Plan, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		emitAnalyzeProgress(request, "full-ai-plan", "disabled", "full-ai mode not requested")
		return fullai.DisabledPlan(analysis.GeneratedAt, options, "full-ai mode not requested"), fullai.DisabledSummary(options, "full-ai mode not requested")
	}

	emitAnalyzeProgress(request, "full-ai-plan", "running", fmt.Sprintf("planning deep exploration with read budget %d and token budget %d", options.ReadBudget, options.TokenBudget))
	plan, summary := s.fullAI.Plan(fullai.Input{
		RootPath:     scanResult.RootPath,
		ScanResult:   scanResult,
		Analysis:     analysis,
		Changes:      changeResult,
		SupportFiles: supportFiles,
	}, options)
	if request.DeterministicOnly {
		note := "deterministic-only remains active; full-ai execution is still planning-only"
		plan.Note = appendFullAINote(plan.Note, note)
		summary.Note = appendFullAINote(summary.Note, note)
	}

	detail := fmt.Sprintf("planned %d evidence target(s) across %d function task(s)", len(plan.Targets), len(plan.Functions))
	if strings.TrimSpace(summary.Note) != "" {
		detail = fmt.Sprintf("%s; %s", detail, summary.Note)
	}
	emitAnalyzeProgress(request, "full-ai-plan", summary.Status, detail)
	return plan, summary
}

func (s Service) collectFullAIEvidence(request AnalyzeRequest, scanResult repo.ScanResult, analysis analyzer.Result, changeResult changes.Result, supportFiles []string, plan fullai.Plan, summary fullai.Summary) (fullai.Evidence, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		return fullai.DisabledEvidence(analysis.GeneratedAt, scanResult.RootPath, options, "full-ai mode not requested"), summary
	}

	emitAnalyzeProgress(request, "full-ai-evidence", "running", fmt.Sprintf("collecting evidence from %d planned target(s)", len(plan.Targets)))
	evidence := s.fullAIEvidence.Collect(fullai.Input{
		RootPath:     scanResult.RootPath,
		ScanResult:   scanResult,
		Analysis:     analysis,
		Changes:      changeResult,
		SupportFiles: supportFiles,
	}, plan)

	summary.Status = "collected"
	summary.CollectedItems = evidence.CollectedItems
	summary.FailedItems = evidence.FailedItems
	if strings.TrimSpace(evidence.Note) != "" {
		summary.Note = appendFullAINote(summary.Note, evidence.Note)
	}

	emitAnalyzeProgress(request, "full-ai-evidence", "collected", fmt.Sprintf("captured %d evidence item(s) with %d read failure(s)", evidence.CollectedItems, evidence.FailedItems))
	return evidence, summary
}

func (s Service) prepareFullAIFunctions(request AnalyzeRequest, scanResult repo.ScanResult, analysis analyzer.Result, changeResult changes.Result, supportFiles []string, plan fullai.Plan, evidence fullai.Evidence, summary fullai.Summary) (fullai.Functions, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		return fullai.DisabledFunctions(analysis.GeneratedAt, options, "full-ai mode not requested"), summary
	}

	emitAnalyzeProgress(request, "full-ai-functions", "running", fmt.Sprintf("preparing function jobs from %d evidence item(s)", evidence.CollectedItems))
	functions := s.fullAIFunctions.Prepare(fullai.Input{
		RootPath:     scanResult.RootPath,
		ScanResult:   scanResult,
		Analysis:     analysis,
		Changes:      changeResult,
		SupportFiles: supportFiles,
	}, plan, evidence)

	summary.Status = "prepared"
	summary.PreparedFunctions = len(functions.Jobs)
	if strings.TrimSpace(functions.Note) != "" {
		summary.Note = appendFullAINote(summary.Note, functions.Note)
	}

	emitAnalyzeProgress(request, "full-ai-functions", "prepared", fmt.Sprintf("prepared %d function job(s)", len(functions.Jobs)))
	return functions, summary
}

func (s Service) executeFullAIFunctions(ctx context.Context, request AnalyzeRequest, analysis analyzer.Result, functions fullai.Functions, evidence fullai.Evidence, summary fullai.Summary) (fullai.Execution, fullai.Summary, error) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		return fullai.DisabledExecution(analysis.GeneratedAt, options, "full-ai mode not requested"), summary, nil
	}
	if request.DeterministicOnly {
		note := "deterministic-only mode enabled; full-ai provider execution skipped"
		execution := fullai.DisabledExecution(analysis.GeneratedAt, options, note)
		execution.Status = "skipped"
		summary.Status = "prepared"
		summary.Note = appendFullAINote(summary.Note, note)
		emitAnalyzeProgress(request, "full-ai-execution", "skipped", note)
		return execution, summary, nil
	}

	providerConfig := s.providerConfig(request)
	if !providerConfig.Enabled() {
		note := "no provider configured; full-ai provider execution disabled"
		execution := fullai.DisabledExecution(analysis.GeneratedAt, options, note)
		summary.Status = "provider-disabled"
		summary.Note = appendFullAINote(summary.Note, note)
		emitAnalyzeProgress(request, "full-ai-execution", "disabled", note)
		return execution, summary, nil
	}
	if err := s.providers.Validate(providerConfig); err != nil {
		note := err.Error()
		execution := fullai.DisabledExecution(analysis.GeneratedAt, options, note)
		execution.Status = "fallback"
		summary.Status = "fallback"
		summary.Note = appendFullAINote(summary.Note, note)
		emitAnalyzeProgress(request, "full-ai-execution", "fallback", note)
		return execution, summary, nil
	}

	client, err := s.providers.ClientFor(providerConfig)
	if err != nil {
		execution := fullai.DisabledExecution(analysis.GeneratedAt, options, err.Error())
		execution.Status = "fallback"
		summary.Status = "fallback"
		summary.Note = appendFullAINote(summary.Note, err.Error())
		emitAnalyzeProgress(request, "full-ai-execution", "fallback", err.Error())
		return execution, summary, nil
	}

	emitAnalyzeProgress(request, "full-ai-execution", "running", fmt.Sprintf("executing %d function job(s) with provider %s", len(functions.Jobs), providerConfig.Name))
	execution := fullai.Execution{
		SchemaVersion: fullai.ExecutionSchemaVersion,
		GeneratedAt:   analysis.GeneratedAt,
		Mode:          string(options.Mode),
		Provider:      strings.TrimSpace(providerConfig.Name),
		Model:         strings.TrimSpace(providerConfig.Model),
		Status:        "running",
		Results:       make([]fullai.FunctionResult, 0, len(functions.Jobs)),
	}

	for _, job := range functions.Jobs {
		result, runErr := executeFullAIJob(ctx, client, providerConfig, job, evidence)
		if runErr != nil {
			if errors.Is(runErr, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				emitAnalyzeProgress(request, "full-ai-execution", "canceled", "full-ai execution canceled")
				return fullai.Execution{}, summary, context.Canceled
			}
			result = fullai.FunctionResult{
				Name:            job.Name,
				Status:          "fallback",
				InstructionPath: job.InstructionPath,
				EvidencePaths:   append([]string(nil), job.EvidencePaths...),
				Error:           runErr.Error(),
			}
		}
		execution.Results = append(execution.Results, result)
		emitAnalyzeProgress(request, "full-ai-function", result.Status, fmt.Sprintf("%s: %s", job.Name, fullAIResultDetail(result)))
	}

	execution.ExecutedCount, execution.VerifiedCount, execution.FailedCount = countFullAIResults(execution.Results)
	execution.Status = fullAIExecutionStatus(execution)
	execution.Note = fmt.Sprintf("executed %d function job(s), verified %d, failed %d", execution.ExecutedCount, execution.VerifiedCount, execution.FailedCount)

	summary.Status = execution.Status
	summary.ExecutedFunctions = execution.ExecutedCount
	summary.VerifiedFunctions = execution.VerifiedCount
	summary.Note = appendFullAINote(summary.Note, execution.Note)
	emitAnalyzeProgress(request, "full-ai-execution", execution.Status, execution.Note)

	return execution, summary, nil
}

func executeFullAIJob(ctx context.Context, client provider.Client, config provider.Config, job fullai.FunctionJob, evidence fullai.Evidence) (fullai.FunctionResult, error) {
	prompt, err := fullai.BuildFunctionPrompt(job, evidence)
	if err != nil {
		return fullai.FunctionResult{}, err
	}

	promptResult, err := client.Complete(ctx, provider.PromptRequest{
		Config:       config,
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   prompt.UserPrompt,
	})
	if err != nil {
		return fullai.FunctionResult{}, err
	}

	output, err := fullai.ParseFunctionOutput(promptResult.Content)
	if err != nil {
		return fullai.FunctionResult{
			Name:            job.Name,
			Status:          "fallback",
			InstructionPath: job.InstructionPath,
			EvidencePaths:   append([]string(nil), job.EvidencePaths...),
			RawOutput:       promptResult.Content,
			Error:           err.Error(),
		}, nil
	}

	output, verified := fullai.VerifyFunctionOutput(job, output)
	status := "verified"
	if !verified {
		status = "unverified"
	}

	return fullai.FunctionResult{
		Name:            job.Name,
		Status:          status,
		InstructionPath: job.InstructionPath,
		EvidencePaths:   append([]string(nil), job.EvidencePaths...),
		RawOutput:       promptResult.Content,
		Output:          output,
		Verified:        verified,
	}, nil
}

func countFullAIResults(results []fullai.FunctionResult) (executed int, verified int, failed int) {
	for _, result := range results {
		switch result.Status {
		case "verified", "unverified":
			executed++
			if result.Verified {
				verified++
			}
		default:
			failed++
		}
	}
	return executed, verified, failed
}

func fullAIExecutionStatus(execution fullai.Execution) string {
	switch {
	case execution.ExecutedCount == 0 && execution.FailedCount > 0:
		return "fallback"
	case execution.FailedCount > 0:
		return "partial"
	case execution.VerifiedCount == execution.ExecutedCount:
		return "executed"
	default:
		return "unverified"
	}
}

func fullAIResultDetail(result fullai.FunctionResult) string {
	if strings.TrimSpace(result.Error) != "" {
		return result.Error
	}
	if strings.TrimSpace(result.Output.Summary) != "" {
		return result.Output.Summary
	}
	return "completed"
}

func fullAICacheStatus(summary fullai.Summary) string {
	if summary.ExecutedFunctions > 0 {
		return cache.StatusMiss
	}
	return cache.StatusDisabled
}

func combineProviderStatus(providerStatus string, fullAIStatus string) string {
	if strings.TrimSpace(fullAIStatus) == "" || fullAIStatus == cache.StatusDisabled {
		return providerStatus
	}
	if strings.TrimSpace(providerStatus) == "" || providerStatus == cache.StatusDisabled {
		return "full-ai " + fullAIStatus
	}
	return providerStatus + ", full-ai " + fullAIStatus
}

func appendFullAINote(existing string, note string) string {
	trimmedExisting := strings.TrimSpace(existing)
	trimmedNote := strings.TrimSpace(note)
	if trimmedExisting == "" {
		return trimmedNote
	}
	if trimmedNote == "" {
		return trimmedExisting
	}
	return trimmedExisting + "; " + trimmedNote
}
