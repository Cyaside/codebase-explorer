package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

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
