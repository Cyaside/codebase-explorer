package analyzer

import (
	"time"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

type Service struct {
	version string
}

const schemaVersion = "phase1.v1"

func NewService(version string) Service {
	return Service{version: version}
}

func (s Service) Analyze(scanResult repo.ScanResult, deterministicOnly bool) Result {
	generatedAt := time.Now().UTC()
	orientationScan := buildOrientationScan(scanResult)
	metrics := buildMetrics(scanResult)
	languages := buildLanguageSummary(orientationScan.Files)
	modules := buildModules(orientationScan.Files)
	hotspots := computeHotspots(orientationScan.Files)
	dependencyRisks := buildDependencyRisks(orientationScan.Files)
	entryPoints := buildEntryPoints(orientationScan.Files)
	importantDirectories := buildImportantDirectories(modules)
	coreModules := buildCoreModules(modules)
	readingPath := buildReadingPath(orientationScan, entryPoints, coreModules, hotspots)
	projectType := guessProjectType(orientationScan, languages, entryPoints)
	summary := buildSummary(scanResult.ProjectName, projectType, metrics, languages, entryPoints, hotspots)

	provider := "deterministic-only"
	if !deterministicOnly {
		provider = "deterministic-baseline"
	}

	return Result{
		SchemaVersion:        schemaVersion,
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
