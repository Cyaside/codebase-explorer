package viewer

import (
	"strings"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
	"github.com/Cyaside/codebase-explorer/internal/graph"
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
			ProviderMode: "compatible",
		},
		Metrics: analyzer.Metrics{
			TotalFiles: 42,
			TotalLines: 1200,
		},
		Warnings: []string{"Large repository detected."},
		Changes: changes.Result{
			Available: true,
			FrequentlyMentionedAreas: []changes.AreaMention{
				{Path: "internal/app", MentionCount: 2, Confidence: changes.ConfidenceStrong},
			},
		},
		AI: provider.Result{
			Status:         provider.ResultStatusSkipped,
			ProjectSummary: "AI summary",
		},
		Graphs: graph.Set{SchemaVersion: graph.SchemaVersion, Views: []graph.View{{
			ID:    "architecture",
			Nodes: []graph.Node{{ID: "entry", Label: "Entry point"}, {ID: "module", Label: "Core module"}},
			Edges: []graph.Edge{{ID: "edge", Source: "entry", Target: "module", Relation: "calls"}},
		}}},
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
	if !strings.Contains(string(files["viewer-data.js"]), "internal/app") {
		t.Fatalf("expected viewer data payload to include change-awareness output")
	}
	if !strings.Contains(string(files["viewer-data.js"]), "Large repository detected.") {
		t.Fatalf("expected viewer data payload to include warnings")
	}
	if !strings.Contains(string(files["viewer-data.js"]), `"graphs"`) || !strings.Contains(string(files["app.js"]), "Open structured graph data") {
		t.Fatal("expected offline viewer to expose structured graphs")
	}
	if strings.Contains(string(files["index.html"]), "Mermaid Source") || strings.Contains(string(files["app.js"]), "Open raw Mermaid file") {
		t.Fatal("new viewer should not advertise Mermaid source")
	}
}
