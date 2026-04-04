package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeWritesChangesCorrelationBundle(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	issuesPath := filepath.Join("..", "..", "testdata", "support-files", "issues.json")
	changelogPath := filepath.Join("..", "..", "testdata", "support-files", "CHANGELOG.md")

	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:             repoPath,
		DeterministicOnly:    true,
		OptionalSupportFiles: []string{issuesPath, changelogPath},
	})
	if err != nil {
		t.Fatalf("analyze with support files: %v", err)
	}

	changesPath := filepath.Join(result.OutputPath, "changes", "issue-correlation.json")
	changesContents, err := os.ReadFile(changesPath)
	if err != nil {
		t.Fatalf("read issue correlation bundle: %v", err)
	}

	var changes struct {
		SchemaVersion            string `json:"schema_version"`
		Available                bool   `json:"available"`
		ParsedItemCount          int    `json:"parsed_item_count"`
		FrequentlyMentionedAreas []struct {
			Path string `json:"path"`
		} `json:"frequently_mentioned_areas"`
	}
	if err := json.Unmarshal(changesContents, &changes); err != nil {
		t.Fatalf("unmarshal issue correlation bundle: %v", err)
	}
	if changes.SchemaVersion != "changes.v1" {
		t.Fatalf("expected changes schema version changes.v1, got %q", changes.SchemaVersion)
	}
	if !changes.Available || changes.ParsedItemCount == 0 {
		t.Fatalf("expected parsed change-awareness output, got %#v", changes)
	}
	if len(changes.FrequentlyMentionedAreas) == 0 {
		t.Fatalf("expected change-aware area mentions to be written")
	}

	readmeContents, err := os.ReadFile(filepath.Join(result.OutputPath, "changes", "README.md"))
	if err != nil {
		t.Fatalf("read changes readme: %v", err)
	}
	if string(readmeContents) == "" {
		t.Fatalf("expected changes readme to be populated")
	}
}
