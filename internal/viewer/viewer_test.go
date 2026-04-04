package viewer

import (
	"strings"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

func TestFilesIncludeViewerAssetsAndData(t *testing.T) {
	t.Parallel()

	files, err := Files(BundleData{
		BundleName:  "bundle-name",
		GeneratedAt: time.Date(2026, time.April, 4, 2, 0, 0, 0, time.UTC),
		Project: ProjectData{
			Name:         "Codebase Explorer",
			Type:         "Go CLI application",
			AnalyzedPath: "/tmp/repo",
			Summary:      "Deterministic summary",
			ProviderMode: "deterministic-only",
		},
		Metrics: analyzer.Metrics{
			TotalFiles: 42,
			TotalLines: 1200,
		},
		AI: provider.Result{
			Status:         provider.ResultStatusSkipped,
			ProjectSummary: "AI summary",
		},
		Mermaid: MermaidData{
			Architecture: "flowchart TD",
			Dependencies: "flowchart LR",
		},
		Links: LinkData{
			Root: "README.md",
		},
	})
	if err != nil {
		t.Fatalf("viewer files: %v", err)
	}

	for _, required := range []string{"index.html", "app.js", "style.css", "viewer-data.js"} {
		if _, found := files[required]; !found {
			t.Fatalf("expected viewer asset %q to be present", required)
		}
	}
	if !strings.Contains(string(files["viewer-data.js"]), "Codebase Explorer") {
		t.Fatalf("expected viewer data payload to include project details")
	}
}
