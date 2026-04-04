package cli

import (
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/app"
)

func TestPrintAnalyzeResultIncludesAISummary(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	printAnalyzeResult(&builder, app.AnalyzeResult{
		ProjectName:     "Codebase Explorer",
		ProjectType:     "Go CLI application",
		TotalFiles:      42,
		TotalLines:      1200,
		EntryPoints:     []string{"cmd/codearch/main.go"},
		PrimaryLanguage: "Go",
		OutputPath:      "/tmp/out",
		Cache: app.AnalyzeCacheSummary{
			Enabled:             true,
			DeterministicStatus: "hit",
			ProviderStatus:      "miss",
		},
		Changes: app.AnalyzeChangesSummary{
			SupportFileCount: 2,
			ParsedItemCount:  5,
			MentionedAreas:   3,
			Note:             "supporting files were parsed successfully",
		},
		AI: app.AnalyzeAISummary{
			Status:         "fallback",
			Provider:       "openai-compatible",
			Model:          "gpt-4.1-mini",
			ContextSummary: "modules=5 hotspots=5 dependencies=5 reading_path=6 (truncated)",
			Note:           "parse synthesis response: invalid character",
		},
		Output: app.AnalyzeOutputSummary{
			RetentionLimit: 10,
			PrunedBundles:  2,
		},
	})

	output := builder.String()
	if !strings.Contains(output, "AI synthesis: fallback") {
		t.Fatalf("expected AI status in analyze output, got %q", output)
	}
	if !strings.Contains(output, "AI provider: openai-compatible (gpt-4.1-mini)") {
		t.Fatalf("expected AI provider line in analyze output, got %q", output)
	}
	if !strings.Contains(output, "AI context: modules=5 hotspots=5 dependencies=5 reading_path=6 (truncated)") {
		t.Fatalf("expected AI context summary in analyze output, got %q", output)
	}
	if !strings.Contains(output, "AI note: parse synthesis response: invalid character") {
		t.Fatalf("expected AI note in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Support files: 2 input(s), 5 parsed item(s)") {
		t.Fatalf("expected support-file summary in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Change awareness: 3 area(s) correlated") {
		t.Fatalf("expected change-awareness summary in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Change note: supporting files were parsed successfully") {
		t.Fatalf("expected change note in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Cache: deterministic hit, provider miss") {
		t.Fatalf("expected cache summary in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Output retention: keep latest 10 bundle(s)") {
		t.Fatalf("expected output retention line in analyze output, got %q", output)
	}
	if !strings.Contains(output, "Output cleanup: removed 2 older bundle(s)") {
		t.Fatalf("expected output cleanup line in analyze output, got %q", output)
	}
}

func TestPrintAnalyzeProgressFormatsEvent(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	printAnalyzeProgress(&builder, app.AnalyzeProgressEvent{
		Stage:  "ai-synthesis",
		Status: "running",
		Detail: "calling provider openai-compatible with model gpt-4.1-mini",
	})

	output := builder.String()
	if !strings.Contains(output, "AI progress [ai-synthesis/running]: calling provider openai-compatible with model gpt-4.1-mini") {
		t.Fatalf("expected progress line to be formatted, got %q", output)
	}
}

func TestPrintOpenResultIncludesResolvedViewerPath(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	printOpenResult(&builder, app.OpenResult{
		BundlePath:     "/tmp/out/bundle",
		ViewerPath:     "/tmp/out/bundle/ui/index.html",
		ResolvedLatest: true,
	}, nil, true)

	output := builder.String()
	if !strings.Contains(output, "Viewer ready.") {
		t.Fatalf("expected open result header, got %q", output)
	}
	if !strings.Contains(output, "Bundle source: latest bundle in output root") {
		t.Fatalf("expected latest bundle note, got %q", output)
	}
	if !strings.Contains(output, "Browser launch: skipped by flag") {
		t.Fatalf("expected browser skip note, got %q", output)
	}
}

func TestPrintCacheClearResult(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	printCacheClearResult(&builder, app.CacheClearResult{
		CacheRoot:      "/tmp/cache",
		RemovedEntries: 3,
	})

	output := builder.String()
	if !strings.Contains(output, "Cache cleared.") || !strings.Contains(output, "Removed entries: 3") {
		t.Fatalf("expected cache clear output, got %q", output)
	}
}
