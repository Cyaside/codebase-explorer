package app

import (
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeWritesDeterministicBundle(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("analyze sample repo: %v", err)
	}

	if result.OutputPath == "" {
		t.Fatalf("expected output path to be returned")
	}
}
