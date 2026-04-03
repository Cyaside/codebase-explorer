package analyzer

import (
	"path"
	"slices"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func buildEntryPoints(files []repo.FileInfo) []string {
	var entryPoints []string
	for _, file := range files {
		if file.IsEntryPoint {
			entryPoints = append(entryPoints, file.Path)
		}
	}
	slices.Sort(entryPoints)
	return entryPoints
}

func buildReadingPath(scanResult repo.ScanResult, entryPoints, coreModules []string, hotspots []Hotspot) []ReadingPathItem {
	seen := map[string]struct{}{}
	var readingPath []ReadingPathItem

	add := func(pathValue, reason string) {
		if pathValue == "" {
			return
		}
		if _, found := seen[pathValue]; found {
			return
		}
		seen[pathValue] = struct{}{}
		readingPath = append(readingPath, ReadingPathItem{
			Path:   pathValue,
			Reason: reason,
		})
	}

	for _, documentationPath := range scanResult.Documentation {
		baseName := path.Base(documentationPath)
		if strings.EqualFold(baseName, "readme.md") || strings.EqualFold(baseName, "readme.txt") {
			add(documentationPath, "start with repository documentation")
		}
	}
	for _, entryPoint := range entryPoints {
		add(entryPoint, "entry point candidate")
	}
	for _, module := range coreModules {
		add(module, "core module by size and structure")
	}
	for _, hotspot := range hotspots {
		add(hotspot.Path, "high-signal file for orientation and risk review")
	}

	if len(readingPath) > 8 {
		readingPath = readingPath[:8]
	}

	return readingPath
}
