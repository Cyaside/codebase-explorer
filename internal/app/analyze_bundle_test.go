package app

import (
	"encoding/json"
	"os"
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
	if result.AI.Status != "skipped" {
		t.Fatalf("expected deterministic-only analyze to mark AI summary as skipped, got %#v", result.AI)
	}

	analysisPath := filepath.Join(result.OutputPath, "data", "analysis.json")
	analysisContents, err := os.ReadFile(analysisPath)
	if err != nil {
		t.Fatalf("read analysis contract: %v", err)
	}

	var analysis struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(analysisContents, &analysis); err != nil {
		t.Fatalf("unmarshal analysis contract: %v", err)
	}
	if analysis.SchemaVersion != "phase1.v1" {
		t.Fatalf("expected schema version phase1.v1, got %q", analysis.SchemaVersion)
	}

	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "contract.json")); err != nil {
		t.Fatalf("expected contract.json to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "full-ai-plan.json")); err != nil {
		t.Fatalf("expected full-ai plan contract to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "full-ai-meta.json")); err != nil {
		t.Fatalf("expected full-ai summary contract to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "full-ai-evidence.json")); err != nil {
		t.Fatalf("expected full-ai evidence contract to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "full-ai-functions.json")); err != nil {
		t.Fatalf("expected full-ai functions contract to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "data", "full-ai-execution.json")); err != nil {
		t.Fatalf("expected full-ai execution contract to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "architecture", "module-graph.mmd")); err != nil {
		t.Fatalf("expected architecture mermaid graph to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "dependencies", "dependency-graph.mmd")); err != nil {
		t.Fatalf("expected dependency mermaid graph to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "ui", "index.html")); err != nil {
		t.Fatalf("expected viewer index to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.OutputPath, "ui", "viewer-data.js")); err != nil {
		t.Fatalf("expected viewer data payload to be written: %v", err)
	}

	aiContextContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-context.json"))
	if err != nil {
		t.Fatalf("read ai context contract: %v", err)
	}

	var aiContext struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(aiContextContents, &aiContext); err != nil {
		t.Fatalf("unmarshal ai context contract: %v", err)
	}
	if aiContext.SchemaVersion != condensedContextSchemaVersion {
		t.Fatalf("expected AI context schema version %q, got %q", condensedContextSchemaVersion, aiContext.SchemaVersion)
	}

	aiResultContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-result.json"))
	if err != nil {
		t.Fatalf("read ai result contract: %v", err)
	}

	var aiResult struct {
		SchemaVersion  string `json:"schema_version"`
		Status         string `json:"status"`
		FallbackReason string `json:"fallback_reason"`
	}
	if err := json.Unmarshal(aiResultContents, &aiResult); err != nil {
		t.Fatalf("unmarshal ai result contract: %v", err)
	}
	if aiResult.SchemaVersion != "ai-result.v1" {
		t.Fatalf("expected AI result schema version ai-result.v1, got %q", aiResult.SchemaVersion)
	}
	if aiResult.Status != "skipped" {
		t.Fatalf("expected deterministic-only analyze to skip AI synthesis, got %#v", aiResult)
	}
}
