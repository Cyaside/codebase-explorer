package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func ReadingPathREADME(analysis analyzer.Result, aiResult provider.Result) string {
	var builder strings.Builder
	builder.WriteString("# Reading Path\n\n")
	builder.WriteString("## Related Views\n\n")
	builder.WriteString("- [Open the local viewer](../ui/index.html)\n")
	builder.WriteString("- [Review the root report](../README.md)\n")
	builder.WriteString("- [Compare with architecture notes](../architecture/README.md)\n\n")
	if len(analysis.ReadingPath) == 0 {
		builder.WriteString("No reading path candidates were generated.\n")
		return builder.String()
	}
	rationales := readingPathRationaleMap(aiResult)
	for index, item := range analysis.ReadingPath {
		builder.WriteString(fmt.Sprintf("%d. `%s` - %s\n", index+1, item.Path, item.Reason))
		if rationale := rationales[item.Path]; rationale != "" {
			builder.WriteString(fmt.Sprintf("   - AI rationale: %s\n", rationale))
		}
	}
	return builder.String()
}
