package app

import (
	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

const (
	condensedContextSchemaVersion = "context.v1"
	maxContextModules             = 5
	maxContextHotspots            = 5
	maxContextDependencies        = 5
	maxContextReadingPath         = 6
)

func buildCondensedContext(analysis analyzer.Result) provider.CondensedContext {
	context := provider.CondensedContext{
		SchemaVersion: condensedContextSchemaVersion,
		GeneratedAt:   analysis.GeneratedAt,
		Project: provider.ProjectFacts{
			Name:            analysis.ProjectName,
			Type:            analysis.ProjectType,
			Summary:         analysis.Summary,
			PrimaryLanguage: primaryLanguage(analysis),
			TotalFiles:      analysis.Metrics.TotalFiles,
			TotalLines:      analysis.Metrics.TotalLines,
		},
		Modules:              limitModules(analysis.Modules),
		EntryPoints:          append([]string(nil), analysis.EntryPoints...),
		Hotspots:             limitHotspots(analysis.Hotspots),
		DependencyHighlights: limitDependencies(analysis.DependencyRisks),
		ReadingPath:          limitReadingPath(analysis.ReadingPath),
	}

	context.Metadata = provider.ContextMetadata{
		ModulesIncluded:      len(context.Modules),
		ModulesTrimmed:       trimmedCount(len(analysis.Modules), len(context.Modules)),
		HotspotsIncluded:     len(context.Hotspots),
		HotspotsTrimmed:      trimmedCount(len(analysis.Hotspots), len(context.Hotspots)),
		DependenciesIncluded: len(context.DependencyHighlights),
		DependenciesTrimmed:  trimmedCount(len(analysis.DependencyRisks), len(context.DependencyHighlights)),
		ReadingPathIncluded:  len(context.ReadingPath),
		ReadingPathTrimmed:   trimmedCount(len(analysis.ReadingPath), len(context.ReadingPath)),
	}
	context.Metadata.Truncated =
		context.Metadata.ModulesTrimmed > 0 ||
			context.Metadata.HotspotsTrimmed > 0 ||
			context.Metadata.DependenciesTrimmed > 0 ||
			context.Metadata.ReadingPathTrimmed > 0

	return context
}

func primaryLanguage(analysis analyzer.Result) string {
	if len(analysis.Languages) == 0 {
		return ""
	}
	return analysis.Languages[0].Name
}

func limitModules(modules []analyzer.ModuleInfo) []provider.ModuleSummary {
	count := min(len(modules), maxContextModules)
	summaries := make([]provider.ModuleSummary, 0, count)
	for _, module := range modules[:count] {
		summaries = append(summaries, provider.ModuleSummary{
			Path:            module.Path,
			FileCount:       module.FileCount,
			TotalLines:      module.TotalLines,
			Languages:       append([]string(nil), module.Languages...),
			EntryPointCount: module.EntryPointCount,
			MarkerCount:     module.MarkerCount,
		})
	}
	return summaries
}

func limitHotspots(hotspots []analyzer.Hotspot) []provider.HotspotSummary {
	count := min(len(hotspots), maxContextHotspots)
	summaries := make([]provider.HotspotSummary, 0, count)
	for _, hotspot := range hotspots[:count] {
		summaries = append(summaries, provider.HotspotSummary{
			Path:    hotspot.Path,
			Score:   hotspot.Score,
			Reasons: append([]string(nil), hotspot.Reasons...),
		})
	}
	return summaries
}

func limitDependencies(risks []analyzer.DependencyRisk) []provider.DependencyHighlight {
	count := min(len(risks), maxContextDependencies)
	summaries := make([]provider.DependencyHighlight, 0, count)
	for _, risk := range risks[:count] {
		summaries = append(summaries, provider.DependencyHighlight{
			Path:        risk.Path,
			ImportCount: risk.ImportCount,
			Reason:      risk.Reason,
		})
	}
	return summaries
}

func limitReadingPath(items []analyzer.ReadingPathItem) []provider.ReadingPathHint {
	count := min(len(items), maxContextReadingPath)
	summaries := make([]provider.ReadingPathHint, 0, count)
	for _, item := range items[:count] {
		summaries = append(summaries, provider.ReadingPathHint{
			Path:   item.Path,
			Reason: item.Reason,
		})
	}
	return summaries
}

func trimmedCount(total, included int) int {
	if total <= included {
		return 0
	}
	return total - included
}
