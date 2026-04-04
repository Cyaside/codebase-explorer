package changes

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

const (
	maxAreaMentions   = 8
	maxUnstableAreas  = 5
	maxThemes         = 8
	maxAreaExamples   = 3
	minThemeMentions  = 2
	minAreaMatchScore = 2
)

type areaCandidate struct {
	Path           string
	NormalizedPath string
	Segments       []string
	IsFile         bool
	RelatedHotspot bool
}

type areaAccumulator struct {
	Path           string
	MentionCount   int
	SourceCount    int
	RelatedHotspot bool
	Examples       []string
	Confidence     string
	reasons        map[string]struct{}
	sources        map[string]struct{}
	maxScore       int
	exactHits      int
}

func buildAreaMentions(analysis analyzer.Result, items []normalizedItem) []AreaMention {
	candidates := buildAreaCandidates(analysis)
	if len(candidates) == 0 || len(items) == 0 {
		return nil
	}

	accumulators := map[string]*areaAccumulator{}
	for _, item := range items {
		for _, candidate := range candidates {
			score, exactMatch, segmentHits := matchCandidate(candidate, item)
			if score < minAreaMatchScore {
				continue
			}

			accumulator, found := accumulators[candidate.Path]
			if !found {
				accumulator = &areaAccumulator{
					Path:           candidate.Path,
					RelatedHotspot: candidate.RelatedHotspot,
					reasons:        map[string]struct{}{},
					sources:        map[string]struct{}{},
				}
				accumulators[candidate.Path] = accumulator
			}

			accumulator.MentionCount++
			accumulator.RelatedHotspot = accumulator.RelatedHotspot || candidate.RelatedHotspot
			accumulator.sources[item.SourcePath] = struct{}{}
			accumulator.SourceCount = len(accumulator.sources)
			if score > accumulator.maxScore {
				accumulator.maxScore = score
			}
			if exactMatch {
				accumulator.exactHits++
				accumulator.reasons["mentioned via explicit path reference"] = struct{}{}
			}
			if segmentHits >= 2 {
				accumulator.reasons["multiple repository terms appear in support notes"] = struct{}{}
			}
			if accumulator.RelatedHotspot {
				accumulator.reasons["also appears in deterministic hotspot candidates"] = struct{}{}
			}
			if title := strings.TrimSpace(item.Title); title != "" && len(accumulator.Examples) < maxAreaExamples && !containsString(accumulator.Examples, title) {
				accumulator.Examples = append(accumulator.Examples, title)
			}
		}
	}

	mentions := make([]AreaMention, 0, len(accumulators))
	for _, accumulator := range accumulators {
		confidence := ConfidenceAmbiguous
		switch {
		case accumulator.exactHits > 0 || accumulator.MentionCount >= 3:
			confidence = ConfidenceStrong
		case accumulator.MentionCount >= 2:
			confidence = ConfidenceModerate
		}
		accumulator.Confidence = confidence

		mentions = append(mentions, AreaMention{
			Path:           accumulator.Path,
			MentionCount:   accumulator.MentionCount,
			SourceCount:    accumulator.SourceCount,
			Confidence:     accumulator.Confidence,
			RelatedHotspot: accumulator.RelatedHotspot,
			Reasons:        sortedKeys(accumulator.reasons),
			Examples:       accumulator.Examples,
		})
	}

	sort.SliceStable(mentions, func(i int, j int) bool {
		if mentions[i].MentionCount != mentions[j].MentionCount {
			return mentions[i].MentionCount > mentions[j].MentionCount
		}
		if mentions[i].RelatedHotspot != mentions[j].RelatedHotspot {
			return mentions[i].RelatedHotspot
		}
		return mentions[i].Path < mentions[j].Path
	})

	return trimAreaMentions(mentions, maxAreaMentions)
}

func buildLikelyUnstableAreas(mentions []AreaMention) []AreaMention {
	candidates := make([]AreaMention, 0, len(mentions))
	for _, mention := range mentions {
		if mention.RelatedHotspot || mention.MentionCount >= 2 {
			candidates = append(candidates, mention)
		}
	}
	return trimAreaMentions(candidates, maxUnstableAreas)
}

func buildHotspotCorrelations(hotspots []analyzer.Hotspot, mentions []AreaMention) []HotspotCorrelation {
	if len(hotspots) == 0 || len(mentions) == 0 {
		return nil
	}

	mentionByPath := make(map[string]AreaMention, len(mentions))
	for _, mention := range mentions {
		mentionByPath[normalizePath(mention.Path)] = mention
	}

	correlations := make([]HotspotCorrelation, 0, len(hotspots))
	for _, hotspot := range hotspots {
		mention, found := mentionByPath[normalizePath(hotspot.Path)]
		if !found {
			continue
		}
		correlations = append(correlations, HotspotCorrelation{
			Path:         hotspot.Path,
			MentionCount: mention.MentionCount,
			Confidence:   mention.Confidence,
		})
	}

	sort.SliceStable(correlations, func(i int, j int) bool {
		if correlations[i].MentionCount != correlations[j].MentionCount {
			return correlations[i].MentionCount > correlations[j].MentionCount
		}
		return correlations[i].Path < correlations[j].Path
	})

	return correlations
}

func buildThemes(items []normalizedItem, mentions []AreaMention) []Theme {
	if len(items) == 0 {
		return nil
	}

	mentionAreas := buildAreaTermIndex(mentions)
	type themeAccumulator struct {
		MentionCount int
		SourcePaths  map[string]struct{}
	}

	accumulators := map[string]*themeAccumulator{}
	for _, item := range items {
		for term := range item.Terms {
			accumulator, found := accumulators[term]
			if !found {
				accumulator = &themeAccumulator{SourcePaths: map[string]struct{}{}}
				accumulators[term] = accumulator
			}
			accumulator.MentionCount++
			accumulator.SourcePaths[item.SourcePath] = struct{}{}
		}
	}

	themes := make([]Theme, 0, len(accumulators))
	for name, accumulator := range accumulators {
		if accumulator.MentionCount < minThemeMentions {
			continue
		}
		themes = append(themes, Theme{
			Name:         name,
			MentionCount: accumulator.MentionCount,
			SourceCount:  len(accumulator.SourcePaths),
			RelatedAreas: mentionAreas[name],
		})
	}

	sort.SliceStable(themes, func(i int, j int) bool {
		if themes[i].MentionCount != themes[j].MentionCount {
			return themes[i].MentionCount > themes[j].MentionCount
		}
		return themes[i].Name < themes[j].Name
	})

	if len(themes) > maxThemes {
		return themes[:maxThemes]
	}
	return themes
}

func buildAreaCandidates(analysis analyzer.Result) []areaCandidate {
	candidateMap := map[string]areaCandidate{}
	addCandidate := func(path string, relatedHotspot bool) {
		normalized := normalizePath(path)
		if normalized == "" {
			return
		}

		segments := areaSegments(path)
		if !allowAreaCandidate(normalized, segments) {
			return
		}

		candidate, found := candidateMap[normalized]
		if !found {
			candidate = areaCandidate{
				Path:           path,
				NormalizedPath: normalized,
				Segments:       segments,
				IsFile:         filepath.Ext(path) != "",
			}
		}
		candidate.RelatedHotspot = candidate.RelatedHotspot || relatedHotspot
		candidateMap[normalized] = candidate
	}

	for _, module := range analysis.Modules {
		addCandidate(module.Path, false)
	}
	for _, file := range analysis.Files {
		parent := normalizePath(filepath.Dir(file.Path))
		if parent == "." || parent == "" {
			continue
		}
		addCandidate(parent, false)
	}
	for _, hotspot := range analysis.Hotspots {
		addCandidate(hotspot.Path, true)
	}

	candidates := make([]areaCandidate, 0, len(candidateMap))
	for _, candidate := range candidateMap {
		candidates = append(candidates, candidate)
	}
	sort.SliceStable(candidates, func(i int, j int) bool {
		return candidates[i].Path < candidates[j].Path
	})
	return candidates
}

func matchCandidate(candidate areaCandidate, item normalizedItem) (int, bool, int) {
	exactMatch := strings.Contains(item.RawText, candidate.NormalizedPath)
	segmentHits := 0
	for _, segment := range candidate.Segments {
		if _, found := item.Terms[segment]; found {
			segmentHits++
		}
	}

	if !exactMatch {
		if candidate.IsFile && segmentHits < 3 {
			return 0, false, segmentHits
		}
		if !candidate.IsFile && segmentHits < 2 {
			return 0, false, segmentHits
		}
	}

	score := 0
	if exactMatch {
		score += 3
	}
	if candidate.IsFile && segmentHits >= 3 {
		score += 2
	}
	if !candidate.IsFile && segmentHits >= 2 {
		score += 2
	}
	return score, exactMatch, segmentHits
}

func areaSegments(path string) []string {
	lowerPath := normalizePath(path)
	parts := strings.FieldsFunc(lowerPath, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSuffix(part, filepath.Ext(part))
		for _, piece := range splitToken(trimmed) {
			if ignoreToken(piece) {
				continue
			}
			if !containsString(segments, piece) {
				segments = append(segments, piece)
			}
		}
	}
	return segments
}

func buildAreaTermIndex(mentions []AreaMention) map[string][]string {
	index := map[string][]string{}
	for _, mention := range mentions {
		for _, segment := range areaSegments(mention.Path) {
			index[segment] = appendUnique(index[segment], mention.Path)
		}
	}
	return index
}

func trimAreaMentions(mentions []AreaMention, maxCount int) []AreaMention {
	if len(mentions) > maxCount {
		return mentions[:maxCount]
	}
	return mentions
}

func normalizePath(path string) string {
	return strings.ToLower(filepath.ToSlash(strings.TrimSpace(path)))
}

func allowAreaCandidate(path string, segments []string) bool {
	if path == "." || len(segments) < 2 {
		return false
	}
	return true
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func appendUnique(values []string, value string) []string {
	if value == "" || containsString(values, value) {
		return values
	}
	return append(values, value)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
