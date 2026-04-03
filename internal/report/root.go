package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func RootREADME(_ repo.ScanResult, analysis analyzer.Result, bundleName string) string {
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
