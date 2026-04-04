package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func ArchitectureREADME(analysis analyzer.Result, aiResult provider.Result) string {
	var builder strings.Builder
	builder.WriteString("# Architecture\n\n")
	if narrative := strings.TrimSpace(aiResult.ArchitectureNarrative); narrative != "" {
		builder.WriteString("## AI Architecture Narrative\n\n")
		builder.WriteString(narrative)
		builder.WriteString("\n\n")
	}
	builder.WriteString("## Orientation Graph\n\n")
	builder.WriteString("- [module-graph.mmd](module-graph.mmd)\n\n")
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
