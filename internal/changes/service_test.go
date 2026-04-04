package changes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestAnalyzeBuildsChangeSignalsFromSupportFiles(t *testing.T) {
	t.Parallel()

	analysis := analyzer.Result{
		GeneratedAt: time.Date(2026, time.April, 4, 4, 0, 0, 0, time.UTC),
		Modules: []analyzer.ModuleInfo{
			{Path: "internal/app"},
			{Path: "internal/core"},
			{Path: "cmd/sample/main.go"},
		},
		Hotspots: []analyzer.Hotspot{
			{Path: "internal/app"},
		},
	}

	result := Analyze(analysis.GeneratedAt, analysis, []string{
		filepath.Join("..", "..", "testdata", "support-files", "issues.json"),
		filepath.Join("..", "..", "testdata", "support-files", "CHANGELOG.md"),
	})

	if !result.Available {
		t.Fatalf("expected change analysis to be available")
	}
	if result.ParsedItemCount < 3 {
		t.Fatalf("expected several parsed items, got %d", result.ParsedItemCount)
	}
	if len(result.FrequentlyMentionedAreas) == 0 {
		t.Fatalf("expected repository areas to be correlated")
	}
	if result.FrequentlyMentionedAreas[0].Path != "internal/app" {
		t.Fatalf("expected internal/app to lead change mentions, got %#v", result.FrequentlyMentionedAreas[0])
	}
	if len(result.HotspotCorrelations) == 0 {
		t.Fatalf("expected hotspot correlation to be recorded")
	}
	if len(result.RepeatedThemes) == 0 {
		t.Fatalf("expected repeated themes to be recorded")
	}
}

func TestAnalyzeKeepsSourceFailuresNonFatal(t *testing.T) {
	t.Parallel()

	analysis := analyzer.Result{
		GeneratedAt: time.Date(2026, time.April, 4, 4, 0, 0, 0, time.UTC),
		Modules: []analyzer.ModuleInfo{
			{Path: "internal/app"},
		},
	}

	result := Analyze(analysis.GeneratedAt, analysis, []string{
		filepath.Join("..", "..", "testdata", "support-files", "issues.json"),
		filepath.Join("..", "..", "testdata", "support-files", "missing.json"),
	})

	if !result.Available {
		t.Fatalf("expected valid support input to keep result available")
	}
	if len(result.Sources) != 2 {
		t.Fatalf("expected two sources to be tracked, got %d", len(result.Sources))
	}
	if result.Sources[1].Status != SourceStatusFailed {
		t.Fatalf("expected missing file to be tracked as failed, got %#v", result.Sources[1])
	}
	if !strings.Contains(result.Note, "could not be parsed") {
		t.Fatalf("expected note to mention skipped source failures, got %q", result.Note)
	}
}

func TestAnalyzeParsesMarkdownIssueNotes(t *testing.T) {
	t.Parallel()

	analysis := analyzer.Result{
		GeneratedAt: time.Date(2026, time.April, 4, 4, 0, 0, 0, time.UTC),
		Files: []repo.FileInfo{
			{Path: "internal/app/app.go"},
			{Path: "internal/core/core.go"},
		},
	}

	result := Analyze(analysis.GeneratedAt, analysis, []string{
		filepath.Join("..", "..", "testdata", "support-files", "issues.md"),
	})

	if !result.Available {
		t.Fatalf("expected markdown issue notes to be parsed")
	}
	if len(result.FrequentlyMentionedAreas) == 0 {
		t.Fatalf("expected markdown issue notes to correlate repository areas")
	}
	if result.FrequentlyMentionedAreas[0].Path != "internal/app" {
		t.Fatalf("expected internal/app correlation, got %#v", result.FrequentlyMentionedAreas[0])
	}
}

func TestAnalyzeMarksMalformedJSONAsFailedSource(t *testing.T) {
	t.Parallel()

	badJSONPath := filepath.Join(t.TempDir(), "broken-issues.json")
	if err := os.WriteFile(badJSONPath, []byte("{ not-valid-json"), 0o644); err != nil {
		t.Fatalf("write malformed support file: %v", err)
	}

	result := Analyze(time.Date(2026, time.April, 4, 4, 0, 0, 0, time.UTC), analyzer.Result{}, []string{badJSONPath})

	if result.Available {
		t.Fatalf("expected malformed-only input to keep change analysis unavailable")
	}
	if len(result.Sources) != 1 || result.Sources[0].Status != SourceStatusFailed {
		t.Fatalf("expected malformed json to be tracked as failed, got %#v", result.Sources)
	}
}
