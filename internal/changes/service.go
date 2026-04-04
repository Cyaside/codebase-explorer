package changes

import (
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func Analyze(generatedAt time.Time, analysis analyzer.Result, supportFiles []string) Result {
	result := Result{
		SchemaVersion:    ResultSchemaVersion,
		GeneratedAt:      generatedAt,
		SupportFileCount: len(supportFiles),
		Sources:          make([]Source, 0, len(supportFiles)),
	}

	if len(supportFiles) == 0 {
		result.Note = "No supporting files were provided, so change correlation was skipped."
		return result
	}

	items := make([]normalizedItem, 0, len(supportFiles))
	for _, supportFile := range supportFiles {
		source, sourceItems := loadSource(supportFile)
		result.Sources = append(result.Sources, source)
		items = append(items, sourceItems...)
	}
	result.ParsedItemCount = len(items)
	if len(items) == 0 {
		result.Note = "Supporting files were provided, but no structured change entries could be extracted."
		return result
	}

	result.Available = true
	result.FrequentlyMentionedAreas = buildAreaMentions(analysis, items)
	result.LikelyUnstableModules = buildLikelyUnstableAreas(result.FrequentlyMentionedAreas)
	result.HotspotCorrelations = buildHotspotCorrelations(analysis.Hotspots, result.FrequentlyMentionedAreas)
	result.RepeatedThemes = buildThemes(items, result.FrequentlyMentionedAreas)

	if len(result.FrequentlyMentionedAreas) == 0 && len(result.RepeatedThemes) == 0 {
		result.Note = "Supporting files were parsed, but no strong repository correlations were detected."
		return result
	}

	if len(result.FrequentlyMentionedAreas) == 0 {
		result.Note = "Supporting files were parsed, but repository path matches remain weak."
		return result
	}

	if hasSourceFailures(result.Sources) {
		result.Note = strings.TrimSpace(result.Note + " Some supporting files could not be parsed and were skipped.")
	}

	return result
}

func hasSourceFailures(sources []Source) bool {
	for _, source := range sources {
		if source.Status == SourceStatusFailed {
			return true
		}
	}
	return false
}
