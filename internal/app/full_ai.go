package app

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
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
