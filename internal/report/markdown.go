package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func RootREADME(scanResult repo.ScanResult, analysis analyzer.Result, bundleName string) string {
	var builder strings.Builder
	builder.WriteString("# Codebase Explorer Report\n\n")
	builder.WriteString(fmt.Sprintf("- Project: `%s`\n", analysis.ProjectName))
	builder.WriteString(fmt.Sprintf("- Bundle: `%s`\n", bundleName))
	builder.WriteString(fmt.Sprintf("- Analyzed path: `%s`\n", analysis.AnalyzedPath))
	builder.WriteString(fmt.Sprintf("- Analysis timestamp: `%s`\n", analysis.GeneratedAt.Format("2006-01-02 15:04:05 MST")))
	builder.WriteString(fmt.Sprintf("- Provider mode: `%s`\n\n", analysis.Provider))
	builder.WriteString("## Summary\n\n")
	builder.WriteString(analysis.Summary)
	builder.WriteString("\n\n## Reports\n\n")
	builder.WriteString("- [Overview](overview/README.md)\n")
	builder.WriteString("- [Architecture](architecture/README.md)\n")
	builder.WriteString("- [Hotspots](hotspots/README.md)\n")
	builder.WriteString("- [Dependencies](dependencies/README.md)\n")
	builder.WriteString("- [Reading Path](reading-path/README.md)\n")
	builder.WriteString("- [Changes](changes/README.md)\n")
	builder.WriteString("- [Data](data/README.md)\n")
	return builder.String()
}

func OverviewREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Overview\n\n")
	builder.WriteString(fmt.Sprintf("Project type: `%s`\n\n", analysis.ProjectType))
	builder.WriteString("## Primary Languages\n\n")
	for _, language := range analysis.Languages {
		builder.WriteString(fmt.Sprintf("- `%s`: %d files, %d lines\n", language.Name, language.FileCount, language.LineCount))
	}
	builder.WriteString("\n## Entry Points\n\n")
	if len(analysis.EntryPoints) == 0 {
		builder.WriteString("- No clear entry point candidate detected.\n")
	} else {
		for _, entryPoint := range analysis.EntryPoints {
			builder.WriteString(fmt.Sprintf("- `%s`\n", entryPoint))
		}
	}
	builder.WriteString("\n## Important Directories\n\n")
	for _, directory := range analysis.ImportantDirectories {
		builder.WriteString(fmt.Sprintf("- `%s`\n", directory))
	}
	builder.WriteString("\n## Deterministic Summary\n\n")
	builder.WriteString(analysis.Summary)
	builder.WriteString("\n")
	return builder.String()
}

func ArchitectureREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Architecture\n\n")
	builder.WriteString("## Core Modules\n\n")
	for _, module := range analysis.Modules {
		builder.WriteString(fmt.Sprintf("- `%s`: %d files, %d lines, %d entry point(s)\n", module.Path, module.FileCount, module.TotalLines, module.EntryPointCount))
	}
	builder.WriteString("\n## Boundary Notes\n\n")
	if len(analysis.CoreModules) == 0 {
		builder.WriteString("- No dominant module boundary detected yet.\n")
	} else {
		builder.WriteString(fmt.Sprintf("- Core reading focus starts with `%s`.\n", strings.Join(analysis.CoreModules, "`, `")))
	}
	return builder.String()
}

func HotspotsREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Hotspots\n\n")
	if len(analysis.Hotspots) == 0 {
		builder.WriteString("No hotspot candidates were detected from the deterministic baseline.\n")
		return builder.String()
	}
	for index, hotspot := range analysis.Hotspots {
		builder.WriteString(fmt.Sprintf("%d. `%s` (score %.2f)\n", index+1, hotspot.Path, hotspot.Score))
		builder.WriteString(fmt.Sprintf("   - lines: %d\n", hotspot.LineCount))
		builder.WriteString(fmt.Sprintf("   - imports: %d\n", hotspot.ImportCount))
		builder.WriteString(fmt.Sprintf("   - markers: %d\n", hotspot.MarkerCount))
		if len(hotspot.Reasons) > 0 {
			builder.WriteString(fmt.Sprintf("   - reasons: %s\n", strings.Join(hotspot.Reasons, "; ")))
		}
	}
	return builder.String()
}

func DependenciesREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Dependencies\n\n")
	if len(analysis.DependencyRisks) == 0 {
		builder.WriteString("No concentrated dependency points were detected from import-like heuristics.\n")
		return builder.String()
	}
	for _, risk := range analysis.DependencyRisks {
		builder.WriteString(fmt.Sprintf("- `%s`: %s\n", risk.Path, risk.Reason))
	}
	return builder.String()
}

func ReadingPathREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Reading Path\n\n")
	if len(analysis.ReadingPath) == 0 {
		builder.WriteString("No reading path candidates were generated.\n")
		return builder.String()
	}
	for index, item := range analysis.ReadingPath {
		builder.WriteString(fmt.Sprintf("%d. `%s` - %s\n", index+1, item.Path, item.Reason))
	}
	return builder.String()
}

func ChangesREADME() string {
	return "# Changes\n\nIssue and changelog correlation is not available in the phase 1 deterministic baseline.\n"
}

func DataREADME() string {
	return "# Data\n\nThis folder contains machine-readable baseline outputs for the deterministic analysis bundle.\n"
}
