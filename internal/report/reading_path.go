package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

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
