package app

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestExportCreatesZipArchiveForLatestBundle(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		CacheRoot:         filepath.Join(outputRoot, "cache"),
		CacheEnabled:      true,
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	analyzeResult, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("seed bundle through analyze: %v", err)
	}

	exportResult, err := service.Export(t.Context(), ExportRequest{})
	if err != nil {
		t.Fatalf("export latest bundle: %v", err)
	}
	if exportResult.BundlePath != analyzeResult.OutputPath {
		t.Fatalf("expected latest bundle to be exported, got %#v", exportResult)
	}
	if _, err := os.Stat(exportResult.ArchivePath); err != nil {
		t.Fatalf("expected archive to be written: %v", err)
	}

	reader, err := zip.OpenReader(exportResult.ArchivePath)
	if err != nil {
		t.Fatalf("open export archive: %v", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		t.Fatalf("expected archive to contain bundle files")
	}
}
