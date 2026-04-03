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
}

func TestDoctorFailsWhenProviderConfigIsInvalid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name: "openai",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	var providerCheck *DoctorCheck
	for index := range result.Checks {
		if result.Checks[index].Name == "provider" {
			providerCheck = &result.Checks[index]
			break
		}
	}
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "fail" {
		t.Fatalf("expected invalid provider config to fail doctor, got %#v", providerCheck)
	}
}

func TestDoctorPassesWhenProviderConfigIsValid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: "https://example.com/v1",
		},
	})

	result, err := service.Doctor(t.Context(), DoctorRequest{})
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}

	var providerCheck *DoctorCheck
	for index := range result.Checks {
		if result.Checks[index].Name == "provider" {
			providerCheck = &result.Checks[index]
			break
		}
	}
	if providerCheck == nil {
		t.Fatalf("expected provider check to be present")
	}
	if providerCheck.Status != "pass" {
		t.Fatalf("expected valid provider config to pass doctor, got %#v", providerCheck)
	}
}
