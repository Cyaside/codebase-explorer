package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func HotspotsREADME(analysis analyzer.Result, aiResult provider.Result, changeResult changes.Result) string {
	var builder strings.Builder
	builder.WriteString("# Hotspots\n\n")
	builder.WriteString("## Related Views\n\n")
	builder.WriteString("- [Open the local viewer](../ui/index.html)\n")
	builder.WriteString("- [Inspect the changes report](../changes/README.md)\n")
	builder.WriteString("- [Review dependency risks](../dependencies/README.md)\n\n")
	if len(analysis.Hotspots) == 0 {
		builder.WriteString("No hotspot candidates were detected from the deterministic baseline.\n")
		return builder.String()
	}
	explanations := hotspotExplanationMap(aiResult)
	changeMentions := hotspotMentionMap(changeResult)
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
		if mention, found := changeMentions[hotspot.Path]; found {
			builder.WriteString(fmt.Sprintf("   - change mentions: %d (%s confidence)\n", mention.MentionCount, mention.Confidence))
		}
	}
	return builder.String()
}

func hotspotMentionMap(result changes.Result) map[string]changes.AreaMention {
	mentions := make(map[string]changes.AreaMention, len(result.FrequentlyMentionedAreas))
	for _, mention := range result.FrequentlyMentionedAreas {
		mentions[mention.Path] = mention
	}
	return mentions
}
