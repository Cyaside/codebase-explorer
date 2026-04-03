package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func HotspotsREADME(analysis analyzer.Result, aiResult provider.Result) string {
	var builder strings.Builder
	builder.WriteString("# Hotspots\n\n")
	if len(analysis.Hotspots) == 0 {
		builder.WriteString("No hotspot candidates were detected from the deterministic baseline.\n")
		return builder.String()
	}
	explanations := hotspotExplanationMap(aiResult)
	for index, hotspot := range analysis.Hotspots {
		builder.WriteString(fmt.Sprintf("%d. `%s` (score %.2f)\n", index+1, hotspot.Path, hotspot.Score))
		builder.WriteString(fmt.Sprintf("   - lines: %d\n", hotspot.LineCount))
		builder.WriteString(fmt.Sprintf("   - imports: %d\n", hotspot.ImportCount))
		builder.WriteString(fmt.Sprintf("   - markers: %d\n", hotspot.MarkerCount))
		if len(hotspot.Reasons) > 0 {
			builder.WriteString(fmt.Sprintf("   - reasons: %s\n", strings.Join(hotspot.Reasons, "; ")))
		}
		if explanation := explanations[hotspot.Path]; explanation != "" {
			builder.WriteString(fmt.Sprintf("   - ai note: %s\n", explanation))
		}
	}
	return builder.String()
}
