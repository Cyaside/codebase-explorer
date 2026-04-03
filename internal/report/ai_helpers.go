package report

import (
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func hotspotExplanationMap(aiResult provider.Result) map[string]string {
	byPath := make(map[string]string, len(aiResult.HotspotExplanations))
	for _, item := range aiResult.HotspotExplanations {
		path := strings.TrimSpace(item.Path)
		explanation := strings.TrimSpace(item.Explanation)
		if path == "" || explanation == "" {
			continue
		}
		byPath[path] = explanation
	}
	return byPath
}

func readingPathRationaleMap(aiResult provider.Result) map[string]string {
	byPath := make(map[string]string, len(aiResult.ReadingPathExplanations))
	for _, item := range aiResult.ReadingPathExplanations {
		path := strings.TrimSpace(item.Path)
		rationale := strings.TrimSpace(item.Rationale)
		if path == "" || rationale == "" {
			continue
		}
		byPath[path] = rationale
	}
	return byPath
}
