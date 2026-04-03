package app

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func emitAnalyzeProgress(request AnalyzeRequest, stage, status, detail string) {
	if request.Progress == nil {
		return
	}

	request.Progress(AnalyzeProgressEvent{
		Stage:  strings.TrimSpace(stage),
		Status: strings.TrimSpace(status),
		Detail: strings.TrimSpace(detail),
	})
}

func summarizeAIContext(context provider.CondensedContext) string {
	summary := fmt.Sprintf(
		"modules=%d hotspots=%d dependencies=%d reading_path=%d",
		len(context.Modules),
		len(context.Hotspots),
		len(context.DependencyHighlights),
		len(context.ReadingPath),
	)
	if context.Metadata.Truncated {
		summary += " (truncated)"
	}
	return summary
}

func buildAISummary(context provider.CondensedContext, result provider.Result) AnalyzeAISummary {
	return AnalyzeAISummary{
		Status:           strings.TrimSpace(result.Status),
		Provider:         strings.TrimSpace(result.Provider),
		Model:            strings.TrimSpace(result.Model),
		Used:             result.Used,
		Note:             strings.TrimSpace(result.FallbackReason),
		ContextSummary:   summarizeAIContext(context),
		ContextTruncated: context.Metadata.Truncated,
	}
}
