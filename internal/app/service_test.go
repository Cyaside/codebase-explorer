package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeWritesDeterministicBundle(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("analyze sample repo: %v", err)
	}

	if result.OutputPath == "" {
		t.Fatalf("expected output path to be returned")
	}

	analysisPath := filepath.Join(result.OutputPath, "data", "analysis.json")
	analysisContents, err := os.ReadFile(analysisPath)
	if err != nil {
		t.Fatalf("read analysis contract: %v", err)
	}

	var analysis struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(analysisContents, &analysis); err != nil {
		t.Fatalf("unmarshal analysis contract: %v", err)
	}
	if analysis.SchemaVersion != "phase1.v1" {
		t.Fatalf("expected schema version phase1.v1, got %q", analysis.SchemaVersion)
	}

	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "contract.json")); err != nil {
		t.Fatalf("expected contract.json to be written: %v", err)
	}

	aiContextContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-context.json"))
	if err != nil {
		t.Fatalf("read ai context contract: %v", err)
	}

	var aiContext struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(aiContextContents, &aiContext); err != nil {
		t.Fatalf("unmarshal ai context contract: %v", err)
	}
	if aiContext.SchemaVersion != condensedContextSchemaVersion {
		t.Fatalf("expected AI context schema version %q, got %q", condensedContextSchemaVersion, aiContext.SchemaVersion)
	}
}

func TestDoctorFailsWhenProviderConfigIsInvalid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name: "openai",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	var providerCheck *DoctorCheck
	for index := range result.Checks {
		if result.Checks[index].Name == "provider" {
			providerCheck = &result.Checks[index]
			break
		}
	}
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "fail" {
		t.Fatalf("expected invalid provider config to fail doctor, got %#v", providerCheck)
	}
}

func TestDoctorPassesWhenProviderConfigIsValid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: "https://example.com/v1",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	var providerCheck *DoctorCheck
	for index := range result.Checks {
		if result.Checks[index].Name == "provider" {
			providerCheck = &result.Checks[index]
			break
		}
	}
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "pass" {
		t.Fatalf("expected valid provider config to pass doctor, got %#v", providerCheck)
	}
}

func TestBuildCondensedContextTrimsLargeSections(t *testing.T) {
	t.Parallel()

	analysis := analyzer.Result{
		GeneratedAt:     testTime(),
		ProjectName:     "example",
		ProjectType:     "Go CLI application",
		Summary:         "example summary",
		Languages:       []analyzer.LanguageSummary{{Name: "Go"}},
		Metrics:         analyzer.Metrics{TotalFiles: 42, TotalLines: 1200},
		EntryPoints:     []string{"cmd/app/main.go"},
		Modules:         makeModules(maxContextModules + 2),
		Hotspots:        makeHotspots(maxContextHotspots + 3),
		DependencyRisks: makeDependencyRisks(maxContextDependencies + 1),
		ReadingPath:     makeReadingPath(maxContextReadingPath + 2),
	}

	context := buildCondensedContext(analysis)

	if len(context.Modules) != maxContextModules {
		t.Fatalf("expected modules to be trimmed to %d, got %d", maxContextModules, len(context.Modules))
	}
	if len(context.Hotspots) != maxContextHotspots {
		t.Fatalf("expected hotspots to be trimmed to %d, got %d", maxContextHotspots, len(context.Hotspots))
	}
	if len(context.DependencyHighlights) != maxContextDependencies {
		t.Fatalf("expected dependencies to be trimmed to %d, got %d", maxContextDependencies, len(context.DependencyHighlights))
	}
	if len(context.ReadingPath) != maxContextReadingPath {
		t.Fatalf("expected reading path to be trimmed to %d, got %d", maxContextReadingPath, len(context.ReadingPath))
	}
	if !context.Metadata.Truncated {
		t.Fatalf("expected truncation metadata to be set")
	}
}

func makeModules(count int) []analyzer.ModuleInfo {
	modules := make([]analyzer.ModuleInfo, 0, count)
	for index := 0; index < count; index++ {
		modules = append(modules, analyzer.ModuleInfo{
			Path:       "module/path",
			FileCount:  index + 1,
			TotalLines: 100 + index,
			Languages:  []string{"Go"},
		})
	}
	return modules
}

func makeHotspots(count int) []analyzer.Hotspot {
	hotspots := make([]analyzer.Hotspot, 0, count)
	for index := 0; index < count; index++ {
		hotspots = append(hotspots, analyzer.Hotspot{
			Path:  "hotspot/path",
			Score: float64(index),
		})
	}
	return hotspots
}

func makeDependencyRisks(count int) []analyzer.DependencyRisk {
	risks := make([]analyzer.DependencyRisk, 0, count)
	for index := 0; index < count; index++ {
		risks = append(risks, analyzer.DependencyRisk{
			Path:        "dependency/path",
			ImportCount: index + 1,
			Reason:      "reason",
		})
	}
	return risks
}

func makeReadingPath(count int) []analyzer.ReadingPathItem {
	items := make([]analyzer.ReadingPathItem, 0, count)
	for index := 0; index < count; index++ {
		items = append(items, analyzer.ReadingPathItem{
			Path:   "reading/path",
			Reason: "reason",
		})
	}
	return items
}

func testTime() time.Time {
	return time.Date(2026, time.April, 3, 15, 0, 0, 0, time.UTC)
}
