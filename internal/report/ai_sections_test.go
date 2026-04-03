package report

import (
	"strings"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestOverviewREADMEIncludesAIProjectSummary(t *testing.T) {
	t.Parallel()

	output := OverviewREADME(testAnalysis(), provider.Result{
		ProjectSummary: "AI summary for the project.",
	})

	if !strings.Contains(output, "## AI Project Summary") {
		t.Fatalf("expected AI summary section to be rendered, got %q", output)
	}
}

func TestArchitectureREADMEIncludesAINarrative(t *testing.T) {
	t.Parallel()

	output := ArchitectureREADME(testAnalysis(), provider.Result{
		ArchitectureNarrative: "The CLI delegates to app orchestration.",
	})

	if !strings.Contains(output, "## AI Architecture Narrative") {
		t.Fatalf("expected AI architecture narrative to be rendered, got %q", output)
	}
}

func TestHotspotsREADMEIncludesMatchingAIExplanation(t *testing.T) {
	t.Parallel()

	output := HotspotsREADME(testAnalysis(), provider.Result{
		HotspotExplanations: []provider.HotspotExplanation{
			{
				Path:        "internal/app/service.go",
				Explanation: "This is the orchestration hotspot.",
			},
		},
	})

	if !strings.Contains(output, "ai note: This is the orchestration hotspot.") {
		t.Fatalf("expected AI hotspot explanation to be rendered, got %q", output)
	}
}

func TestReadingPathREADMEIncludesMatchingAIRationale(t *testing.T) {
	t.Parallel()

	output := ReadingPathREADME(testAnalysis(), provider.Result{
		ReadingPathExplanations: []provider.ReadingPathExplanation{
			{
				Path:      "cmd/codearch/main.go",
				Rationale: "Bootstrap starts here.",
			},
		},
	})

	if !strings.Contains(output, "AI rationale: Bootstrap starts here.") {
		t.Fatalf("expected AI reading path rationale to be rendered, got %q", output)
	}
}

func TestRootREADMEIncludesAISnapshot(t *testing.T) {
	t.Parallel()

	output := RootREADME(repo.ScanResult{}, testAnalysis(), provider.Result{
		Status:         provider.ResultStatusAvailable,
		ProjectSummary: "AI snapshot",
	}, "bundle-name")

	if !strings.Contains(output, "## AI Snapshot") {
		t.Fatalf("expected AI snapshot section to be rendered, got %q", output)
	}
}

func testAnalysis() analyzer.Result {
	return analyzer.Result{
		GeneratedAt:  time.Date(2026, time.April, 3, 16, 30, 0, 0, time.UTC),
		ProjectName:  "Codebase Explorer",
		AnalyzedPath: "/tmp/codebase-explorer",
		ProjectType:  "Go CLI application",
		Summary:      "Deterministic summary.",
		Provider:     "deterministic-only",
		Languages: []analyzer.LanguageSummary{
			{Name: "Go", FileCount: 12, LineCount: 1200},
		},
		EntryPoints:          []string{"cmd/codearch/main.go"},
		ImportantDirectories: []string{"cmd", "internal"},
		CoreModules:          []string{"internal/app"},
		Modules: []analyzer.ModuleInfo{
			{
				Path:            "internal/app",
				FileCount:       3,
				TotalLines:      240,
				EntryPointCount: 0,
			},
		},
		Hotspots: []analyzer.Hotspot{
			{
				Path:        "internal/app/service.go",
				Score:       8.2,
				LineCount:   159,
				ImportCount: 7,
				MarkerCount: 1,
				Reasons:     []string{"high coordination load"},
			},
		},
		ReadingPath: []analyzer.ReadingPathItem{
			{
				Path:   "cmd/codearch/main.go",
				Reason: "Entry point",
			},
		},
	}
}
