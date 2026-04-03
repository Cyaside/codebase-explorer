package analyzer

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func buildMetrics(scanResult repo.ScanResult) Metrics {
	metrics := Metrics{
		TotalFiles:       len(scanResult.Files),
		TotalDirectories: len(scanResult.Directories),
	}
	for _, file := range scanResult.Files {
		metrics.TotalLines += file.LineCount
		metrics.TodoCount += file.TodoCount
		metrics.FixmeCount += file.FixmeCount
		metrics.HackCount += file.HackCount
	}
	return metrics
}

func buildLanguageSummary(files []repo.FileInfo) []LanguageSummary {
	type aggregate struct {
		fileCount int
		lineCount int
	}

	languages := map[string]aggregate{}
	for _, file := range files {
		language := detectLanguage(file.Extension)
		current := languages[language]
		current.fileCount++
		current.lineCount += file.LineCount
		languages[language] = current
	}

	summary := make([]LanguageSummary, 0, len(languages))
	for name, aggregate := range languages {
		summary = append(summary, LanguageSummary{
			Name:      name,
			FileCount: aggregate.fileCount,
			LineCount: aggregate.lineCount,
		})
	}

	sort.Slice(summary, func(i, j int) bool {
		if summary[i].LineCount == summary[j].LineCount {
			return summary[i].Name < summary[j].Name
		}
		return summary[i].LineCount > summary[j].LineCount
	})

	return summary
}

func buildDependencyRisks(files []repo.FileInfo) []DependencyRisk {
	var risks []DependencyRisk
	for _, file := range files {
		if file.ImportCount == 0 {
			continue
		}

		reason := fmt.Sprintf("%d import-like statements suggest a concentrated dependency surface", file.ImportCount)
		risks = append(risks, DependencyRisk{
			Path:        file.Path,
			ImportCount: file.ImportCount,
			Reason:      reason,
		})
	}

	sort.Slice(risks, func(i, j int) bool {
		if risks[i].ImportCount == risks[j].ImportCount {
			return risks[i].Path < risks[j].Path
		}
		return risks[i].ImportCount > risks[j].ImportCount
	})

	if len(risks) > 8 {
		risks = risks[:8]
	}

	return risks
}

func guessProjectType(scanResult repo.ScanResult, languages []LanguageSummary, entryPoints []string) string {
	manifestSet := map[string]struct{}{}
	for _, manifest := range scanResult.ManifestFiles {
		manifestSet[strings.ToLower(path.Base(manifest))] = struct{}{}
	}

	_, hasGoMod := manifestSet["go.mod"]
	_, hasPackageJSON := manifestSet["package.json"]

	switch {
	case hasGoMod && len(entryPoints) > 0:
		return "Go CLI application"
	case hasGoMod:
		return "Go project"
	case hasPackageJSON:
		return "JavaScript or TypeScript project"
	case len(languages) > 0:
		return fmt.Sprintf("%s codebase", languages[0].Name)
	default:
		return "general source repository"
	}
}

func buildSummary(projectName, projectType string, metrics Metrics, languages []LanguageSummary, entryPoints []string, hotspots []Hotspot) string {
	parts := []string{
		fmt.Sprintf("%s appears to be a %s", projectName, strings.ToLower(projectType)),
		fmt.Sprintf("with %d files and %d lines analyzed", metrics.TotalFiles, metrics.TotalLines),
	}
	if len(languages) > 0 {
		parts = append(parts, fmt.Sprintf("primary language: %s", languages[0].Name))
	}
	if len(entryPoints) > 0 {
		parts = append(parts, fmt.Sprintf("entry point focus starts at %s", entryPoints[0]))
	}
	if len(hotspots) > 0 {
		parts = append(parts, fmt.Sprintf("top hotspot: %s", hotspots[0].Path))
	}

	return strings.Join(parts, "; ") + "."
}
