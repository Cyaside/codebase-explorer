package report

import (
	"fmt"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/changes"
)

func ChangesREADME(result changes.Result) string {
	var builder strings.Builder
	builder.WriteString("# Changes\n\n")
	builder.WriteString("## Related Views\n\n")
	builder.WriteString("- [Open the local viewer](../ui/index.html)\n")
	builder.WriteString("- [Review hotspot candidates](../hotspots/README.md)\n")
	builder.WriteString("- [Inspect the raw change-correlation data](issue-correlation.json)\n\n")

	builder.WriteString("## Support Files\n\n")
	if len(result.Sources) == 0 {
		builder.WriteString("- No supporting files were provided.\n")
	} else {
		for _, source := range result.Sources {
			builder.WriteString(fmt.Sprintf("- `%s` (%s/%s): `%s`", source.Path, source.Kind, source.Format, source.Status))
			if source.ItemCount > 0 {
				builder.WriteString(fmt.Sprintf(" with %d parsed item(s)", source.ItemCount))
			}
			if source.Message != "" {
				builder.WriteString(fmt.Sprintf(" - %s", source.Message))
			}
			builder.WriteString("\n")
		}
	}

	if note := strings.TrimSpace(result.Note); note != "" {
		builder.WriteString("\n## Status\n\n")
		builder.WriteString(note)
		builder.WriteString("\n")
	}

	builder.WriteString("\n## Frequently Mentioned Areas\n\n")
	if len(result.FrequentlyMentionedAreas) == 0 {
		builder.WriteString("- No repository areas were matched strongly enough.\n")
	} else {
		for _, area := range result.FrequentlyMentionedAreas {
			builder.WriteString(fmt.Sprintf("- `%s`: %d mention(s), confidence `%s`\n", area.Path, area.MentionCount, area.Confidence))
		}
	}

	builder.WriteString("\n## Repeated Themes\n\n")
	if len(result.RepeatedThemes) == 0 {
		builder.WriteString("- No repeated themes met the reporting threshold.\n")
	} else {
		for _, theme := range result.RepeatedThemes {
			builder.WriteString(fmt.Sprintf("- `%s`: %d mention(s) across %d source(s)\n", theme.Name, theme.MentionCount, theme.SourceCount))
		}
	}

	builder.WriteString("\n## Likely Unstable Modules\n\n")
	if len(result.LikelyUnstableModules) == 0 {
		builder.WriteString("- No modules crossed the instability threshold.\n")
	} else {
		for _, area := range result.LikelyUnstableModules {
			builder.WriteString(fmt.Sprintf("- `%s`: %d mention(s), confidence `%s`\n", area.Path, area.MentionCount, area.Confidence))
		}
	}

	return builder.String()
}
