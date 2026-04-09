package fullai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestCollectorReadsExactAndModuleTargets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "sample"), 0o755); err != nil {
		t.Fatalf("create sample tree: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "app"), 0o755); err != nil {
		t.Fatalf("create internal tree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "sample", "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "app", "service.go"), []byte("package app\nfunc Run() {}\n"), 0o644); err != nil {
		t.Fatalf("write service file: %v", err)
	}

	collector := NewCollector()
	evidence := collector.Collect(Input{
		RootPath: root,
		ScanResult: repo.ScanResult{
			Files: []repo.FileInfo{
				{Path: "cmd/sample/main.go", LineCount: 2, IsEntryPoint: true},
				{Path: "internal/app/service.go", LineCount: 2},
			},
		},
		Analysis: analyzer.Result{
			GeneratedAt: time.Date(2026, time.April, 10, 2, 0, 0, 0, time.UTC),
		},
	}, Plan{
		Mode:       string(ModeFull),
		ReadBudget: 2,
		Targets: []Target{
			{Path: "cmd/sample/main.go", Reason: "entry point", Source: "reading-path", Priority: 100},
			{Path: "internal/app", Reason: "module", Source: "module", Priority: 80},
		},
	})

	if evidence.CollectedItems != 2 {
		t.Fatalf("expected two evidence items, got %#v", evidence)
	}
	if evidence.Items[0].ReadStatus != "read" || evidence.Items[0].Resolution != "exact-file" {
		t.Fatalf("expected exact file read, got %#v", evidence.Items[0])
	}
	if evidence.Items[1].ReadStatus != "read" || evidence.Items[1].Resolution != "module-representative" {
		t.Fatalf("expected module representative read, got %#v", evidence.Items[1])
	}
}
