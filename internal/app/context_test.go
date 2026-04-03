package app

import (
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

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
