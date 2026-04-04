package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func DependenciesREADME(analysis analyzer.Result) string {
	var builder strings.Builder
	builder.WriteString("# Dependencies\n\n")
	builder.WriteString("## Dependency Graph\n\n")
	builder.WriteString("- [dependency-graph.mmd](dependency-graph.mmd)\n\n")
	if len(analysis.DependencyRisks) == 0 {
		builder.WriteString("No concentrated dependency points were detected from import-like heuristics.\n")
		return builder.String()
	}
	for _, risk := range analysis.DependencyRisks {
		builder.WriteString(fmt.Sprintf("- `%s`: %s\n", risk.Path, risk.Reason))
	}
	return builder.String()
}
