package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/cache"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func (s Service) buildFullAIPlan(request AnalyzeRequest, scanResult repo.ScanResult, analysis analyzer.Result, changeResult changes.Result, supportFiles []string) (fullai.Plan, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		emitAnalyzeProgress(request, "full-ai-plan", "disabled", "full-ai mode not requested")
		return fullai.DisabledPlan(analysis.GeneratedAt, options, "full-ai mode not requested"), fullai.DisabledSummary(options, "full-ai mode not requested")
	}

	emitAnalyzeProgress(request, "full-ai-plan", "running", "selecting repository evidence for analysis")
	plan, summary := s.fullAI.Plan(fullai.Input{
		RootPath:     scanResult.RootPath,
		ScanResult:   scanResult,
		Analysis:     analysis,
		Changes:      changeResult,
		SupportFiles: supportFiles,
	}, options)
	detail := fmt.Sprintf("selected %d evidence target(s) for %d analysis section(s)", len(plan.Targets), len(plan.Functions))
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

	bytes := 0
	sources := map[string]int{}
	for _, item := range evidence.Items {
		if item.ReadStatus == "read" {
			bytes += len(item.Snippet)
			sources[item.Source]++
		}
	}
	parts := make([]string, 0, len(sources))
	for source, count := range sources {
		parts = append(parts, fmt.Sprintf("%s:%d", source, count))
	}
	slices.Sort(parts)
	emitAnalyzeProgress(request, "full-ai-evidence", "collected", fmt.Sprintf("selected %d evidence item(s), approximately %d excerpt bytes for AI (%s); %d unreadable or skipped", evidence.CollectedItems-evidence.FailedItems, bytes, strings.Join(parts, ", "), evidence.FailedItems))
	return evidence, summary
}

func (s Service) prepareFullAIFunctions(request AnalyzeRequest, scanResult repo.ScanResult, analysis analyzer.Result, changeResult changes.Result, supportFiles []string, plan fullai.Plan, evidence fullai.Evidence, summary fullai.Summary) (fullai.Functions, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		return fullai.DisabledFunctions(analysis.GeneratedAt, options, "full-ai mode not requested"), summary
	}

	emitAnalyzeProgress(request, "full-ai-functions", "running", fmt.Sprintf("preparing analysis from %d evidence item(s)", evidence.CollectedItems))
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

	emitAnalyzeProgress(request, "full-ai-functions", "prepared", fmt.Sprintf("prepared %d analysis section(s)", len(functions.Jobs)))
	return functions, summary
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
	case execution.VerifiedCount == 0:
		return "failed"
	case execution.VerifiedCount < len(execution.Results):
		return "partial"
	case execution.VerifiedCount > 0:
		return "executed"
	default:
		return "failed"
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
	if summary.CacheStatus != "" {
		return summary.CacheStatus
	}
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
