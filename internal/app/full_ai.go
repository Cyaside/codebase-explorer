package app

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
)

func (s Service) buildFullAIPlan(request AnalyzeRequest, analysis analyzer.Result, changeResult changes.Result, supportFiles []string) (fullai.Plan, fullai.Summary) {
	options := request.FullAI.Normalize()
	if !options.Mode.Enabled() {
		emitAnalyzeProgress(request, "full-ai-plan", "disabled", "full-ai mode not requested")
		return fullai.DisabledPlan(analysis.GeneratedAt, options, "full-ai mode not requested"), fullai.DisabledSummary(options, "full-ai mode not requested")
	}

	emitAnalyzeProgress(request, "full-ai-plan", "running", fmt.Sprintf("planning deep exploration with read budget %d and token budget %d", options.ReadBudget, options.TokenBudget))
	plan, summary := s.fullAI.Plan(fullai.Input{
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
