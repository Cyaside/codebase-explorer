package analyzer

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type Service struct {
	version string
}

func NewService(version string) Service {
	return Service{version: version}
}

func (s Service) Analyze(scanResult repo.ScanResult, deterministicOnly bool) Result {
	generatedAt := time.Now().UTC()
	metrics := buildMetrics(scanResult)
	languages := buildLanguageSummary(scanResult.Files)
	modules := buildModules(scanResult.Files)
	hotspots := computeHotspots(scanResult.Files)
	dependencyRisks := buildDependencyRisks(scanResult.Files)
	entryPoints := buildEntryPoints(scanResult.Files)
	importantDirectories := buildImportantDirectories(modules)
	coreModules := buildCoreModules(modules)
	readingPath := buildReadingPath(scanResult, entryPoints, coreModules, hotspots)
	projectType := guessProjectType(scanResult, languages, entryPoints)
	summary := buildSummary(scanResult.ProjectName, projectType, metrics, languages, entryPoints, hotspots)

	provider := "deterministic-only"
	if !deterministicOnly {
		provider = "deterministic-baseline"
	}

	return Result{
		Version:              s.version,
		GeneratedAt:          generatedAt,
		ProjectName:          scanResult.ProjectName,
		AnalyzedPath:         scanResult.RootPath,
		ProjectType:          projectType,
		Summary:              summary,
		Provider:             provider,
		Languages:            languages,
		ImportantDirectories: importantDirectories,
		EntryPoints:          entryPoints,
		CoreModules:          coreModules,
		Hotspots:             hotspots,
		DependencyRisks:      dependencyRisks,
		ReadingPath:          readingPath,
		Modules:              modules,
		Metrics:              metrics,
		Files:                scanResult.Files,
	}
}

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

func buildModules(files []repo.FileInfo) []ModuleInfo {
	type aggregate struct {
		fileCount       int
		totalLines      int
		entryPointCount int
		markerCount     int
		languages       map[string]struct{}
	}

	modules := map[string]aggregate{}
	for _, file := range files {
		modulePath := topLevelPath(file.Path)
		current := modules[modulePath]
		if current.languages == nil {
			current.languages = map[string]struct{}{}
		}
		current.fileCount++
		current.totalLines += file.LineCount
		current.markerCount += markerTotal(file)
		if file.IsEntryPoint {
			current.entryPointCount++
		}
		current.languages[detectLanguage(file.Extension)] = struct{}{}
		modules[modulePath] = current
	}

	results := make([]ModuleInfo, 0, len(modules))
	for modulePath, aggregate := range modules {
		languages := make([]string, 0, len(aggregate.languages))
		for language := range aggregate.languages {
			languages = append(languages, language)
		}
		slices.Sort(languages)

		results = append(results, ModuleInfo{
			Path:            modulePath,
			FileCount:       aggregate.fileCount,
			TotalLines:      aggregate.totalLines,
			Languages:       languages,
			EntryPointCount: aggregate.entryPointCount,
			MarkerCount:     aggregate.markerCount,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalLines == results[j].TotalLines {
			return results[i].Path < results[j].Path
		}
		return results[i].TotalLines > results[j].TotalLines
	})
	return results
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

func buildImportantDirectories(modules []ModuleInfo) []string {
	var results []string
	for _, module := range modules {
		if module.Path == "." {
			continue
		}
		results = append(results, module.Path)
		if len(results) == 5 {
			break
		}
	}
	return results
}

func buildCoreModules(modules []ModuleInfo) []string {
	var results []string
	for _, module := range modules {
		if module.Path == "." {
			continue
		}
		results = append(results, module.Path)
		if len(results) == 4 {
			break
		}
	}
	return results
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
		if strings.EqualFold(path.Base(documentationPath), "readme.md") || strings.EqualFold(path.Base(documentationPath), "readme.txt") {
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

func topLevelPath(filePath string) string {
	normalized := filepathToSlash(filePath)
	if normalized == "." || normalized == "" {
		return "."
	}
	if !strings.Contains(normalized, "/") {
		return "."
	}
	segments := strings.Split(normalized, "/")
	if len(segments) == 0 {
		return "."
	}
	return segments[0]
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}

func sortHotspots(hotspots []Hotspot) {
	sort.Slice(hotspots, func(i, j int) bool {
		if hotspots[i].Score == hotspots[j].Score {
			return hotspots[i].Path < hotspots[j].Path
		}
		return hotspots[i].Score > hotspots[j].Score
	})
}
