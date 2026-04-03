package analyzer

import (
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestAnalyzeDetectsLanguagesAndHotspots(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "sample-repo")
	scanResult, err := repo.NewScanner().Scan(t.Context(), repo.ScanOptions{RootPath: root})
	if err != nil {
		t.Fatalf("scan sample repo: %v", err)
	}

	result := NewService("test").Analyze(scanResult, true)

	if result.ProjectType == "" {
		t.Fatalf("expected project type to be detected")
	}
	if len(result.Languages) == 0 || result.Languages[0].Name != "Go" {
		t.Fatalf("expected Go to be detected as the primary language, got %#v", result.Languages)
	}
	if len(result.EntryPoints) == 0 {
		t.Fatalf("expected at least one entry point")
	}
	if len(result.Hotspots) == 0 {
		t.Fatalf("expected hotspot candidates")
	}
}
