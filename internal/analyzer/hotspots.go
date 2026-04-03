package analyzer

import (
	"fmt"
	"math"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func computeHotspots(files []repo.FileInfo) []Hotspot {
	maxLines := 1
	maxImports := 1
	maxMarkers := 1
	for _, file := range files {
		maxLines = max(maxLines, file.LineCount)
		maxImports = max(maxImports, file.ImportCount)
		markers := markerTotal(file)
		maxMarkers = max(maxMarkers, markers)
	}

	hotspots := make([]Hotspot, 0, len(files))
	for _, file := range files {
		if file.LineCount == 0 && file.ImportCount == 0 && markerTotal(file) == 0 {
			continue
		}

		score := 0.0
		reasons := make([]string, 0, 4)

		lineRatio := float64(file.LineCount) / float64(maxLines)
		if lineRatio > 0 {
			score += lineRatio * 0.5
		}
		if lineRatio >= 0.6 {
			reasons = append(reasons, fmt.Sprintf("large file (%d lines)", file.LineCount))
		}

		importRatio := float64(file.ImportCount) / float64(maxImports)
		if importRatio > 0 {
			score += importRatio * 0.3
		}
		if importRatio >= 0.5 && file.ImportCount > 0 {
			reasons = append(reasons, fmt.Sprintf("high import activity (%d imports)", file.ImportCount))
		}

		markerCount := markerTotal(file)
		markerRatio := float64(markerCount) / float64(maxMarkers)
		if markerRatio > 0 {
			score += markerRatio * 0.2
		}
		if markerCount > 0 {
			reasons = append(reasons, fmt.Sprintf("attention markers (%d)", markerCount))
		}

		if file.IsEntryPoint {
			score += 0.1
			reasons = append(reasons, "entry point candidate")
		}

		hotspots = append(hotspots, Hotspot{
			Path:        file.Path,
			Score:       math.Round(score*100) / 100,
			Reasons:     reasons,
			LineCount:   file.LineCount,
			ImportCount: file.ImportCount,
			MarkerCount: markerCount,
		})
	}

	sortHotspots(hotspots)
	if len(hotspots) > 10 {
		hotspots = hotspots[:10]
	}
	return hotspots
}

func markerTotal(file repo.FileInfo) int {
	return file.TodoCount + file.FixmeCount + file.HackCount
}
