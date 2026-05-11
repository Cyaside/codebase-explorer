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
		Files:       []repo.FileInfo{{Path: "cmd/main.go"}, {Path: "internal/core/core.go"}, {Path: "ui/main.ts"}, {Path: "ui/lib/run.ts"}, {Path: "ui/src/app.ts"}, {Path: "ui/src/lib/run.ts"}},
	}
	execution := fullai.Execution{Results: []fullai.FunctionResult{{Name: "flowchart", Verified: true, Output: fullai.FunctionOutput{
		GraphEdges: []fullai.GraphEdge{{From: "CLI", To: "Core", EvidencePaths: []string{"cmd/main.go"}}},
	}}}}
	graphs := Build(root, analysis, execution)
	if graphs.SchemaVersion != SchemaVersion || len(graphs.Views) != 3 {
		t.Fatalf("invalid graph set: %#v", graphs)
	}
	if len(graphs.Views[0].Edges) == 0 || len(graphs.Views[1].Edges) != 1 || len(graphs.Views[2].Edges) != 3 {
		t.Fatalf("expected entry, flow, and three resolved import edges: %#v", graphs.Views)
	}
	for _, edge := range graphs.Views[2].Edges {
		if edge.SourceKind != "parsed-import" || len(edge.EvidencePaths) != 2 {
			t.Fatalf("dependency edge must cite source and target: %#v", edge)
		}
	}
}
