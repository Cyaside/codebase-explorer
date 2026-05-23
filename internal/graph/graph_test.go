package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestBuildUsesResolvedImportsAndEvidenceBackedFlow(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":                "module example.com/demo\n",
		"cmd/main.go":           "package main\nimport _ \"example.com/demo/internal/core\"\n",
		"cmd/second.go":         "package main\nimport _ \"example.com/demo/internal/core\"\n",
		"internal/core/core.go": "package core\n",
		"ui/main.ts":            "import { run } from './lib/run'\n",
		"ui/lib/run.ts":         "export const run = () => 1\n",
		"ui/src/app.ts":         "import { run } from '@/lib/run'\n",
		"ui/src/lib/run.ts":     "export const run = () => 2\n",
	} {
		filePath := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	analysis := analyzer.Result{
		Modules:     []analyzer.ModuleInfo{{Path: "cmd"}, {Path: "internal/core"}},
		EntryPoints: []string{"cmd/main.go"},
		Files:       []repo.FileInfo{{Path: "cmd/main.go"}, {Path: "cmd/second.go"}, {Path: "internal/core/core.go"}, {Path: "ui/main.ts"}, {Path: "ui/lib/run.ts"}, {Path: "ui/src/app.ts"}, {Path: "ui/src/lib/run.ts"}},
	}
	execution := fullai.Execution{Results: []fullai.FunctionResult{
		{Name: "architecture", Verified: true, Output: fullai.FunctionOutput{
			GraphEdges: []fullai.GraphEdge{{From: "cmd/main.go", To: "Core", Label: "starts", EvidencePaths: []string{"cmd/main.go"}}},
		}},
		{Name: "flowchart", Verified: true, Output: fullai.FunctionOutput{
			GraphEdges: []fullai.GraphEdge{{From: "CLI", To: "Core", EvidencePaths: []string{"cmd/main.go"}}},
		}},
	}}
	graphs := Build(root, analysis, execution)
	if graphs.SchemaVersion != SchemaVersion || len(graphs.Views) != 3 {
		t.Fatalf("invalid graph set: %#v", graphs)
	}
	if len(graphs.Views[0].Edges) != 2 || len(graphs.Views[1].Edges) != 1 || len(graphs.Views[2].Edges) != 3 {
		t.Fatalf("expected entry and AI architecture, flow, and three resolved import edges: %#v", graphs.Views)
	}
	if graphs.Views[0].Edges[1].SourceKind != "ai-evidence" && graphs.Views[0].Edges[0].SourceKind != "ai-evidence" {
		t.Fatalf("AI architecture edge was discarded: %#v", graphs.Views[0].Edges)
	}
	entryNodes := 0
	for _, node := range graphs.Views[0].Nodes {
		if node.Path == "cmd/main.go" {
			entryNodes++
		}
	}
	if entryNodes != 1 {
		t.Fatalf("entry point was duplicated by AI architecture output: %#v", graphs.Views[0].Nodes)
	}
	for _, edge := range graphs.Views[2].Edges {
		if edge.SourceKind != "parsed-import" {
			t.Fatalf("dependency edge must come from a parsed import: %#v", edge)
		}
		wantEvidence := 2
		if edge.Source == nodeID("dependency", "cmd") {
			wantEvidence = 3
		}
		if len(edge.EvidencePaths) != wantEvidence {
			t.Fatalf("dependency edge must cite all import sources and a target: %#v", edge)
		}
	}
	for _, node := range graphs.Views[2].Nodes {
		if len(node.EvidencePaths) == 0 {
			t.Fatalf("dependency module has no file evidence: %#v", node)
		}
	}
}
