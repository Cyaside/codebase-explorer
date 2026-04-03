package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeFallsBackWhenProviderConfigIsInvalid(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name: "openai",
		},
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: repoPath,
	})
	if err != nil {
		t.Fatalf("analyze sample repo with invalid provider config: %v", err)
	}

	if result.AI.Status != "fallback" {
		t.Fatalf("expected analyze result to surface fallback status, got %#v", result.AI)
	}
	if result.AI.Note == "" {
		t.Fatalf("expected analyze result to surface fallback note")
	}

	aiResultContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-result.json"))
	if err != nil {
		t.Fatalf("read ai result contract: %v", err)
	}

	var aiResult struct {
		Status         string `json:"status"`
		FallbackReason string `json:"fallback_reason"`
	}
	if err := json.Unmarshal(aiResultContents, &aiResult); err != nil {
		t.Fatalf("unmarshal ai result contract: %v", err)
	}
	if aiResult.Status != "fallback" {
		t.Fatalf("expected invalid provider config to fall back, got %#v", aiResult)
	}
	if aiResult.FallbackReason == "" {
		t.Fatalf("expected fallback reason to be present")
	}
}

func TestAnalyzeUsesConfiguredProviderWhenAvailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("expected chat completions path, got %q", r.URL.Path)
		}
		if authorization := r.Header.Get("Authorization"); authorization != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", authorization)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"project_summary\":\"AI summary\",\"architecture_narrative\":\"AI narrative\",\"hotspot_explanations\":[{\"path\":\"internal/app/service.go\",\"explanation\":\"Coordinates the analyze flow.\"}],\"reading_path_explanations\":[{\"path\":\"cmd/codearch/main.go\",\"rationale\":\"Shows application startup.\"}]}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: repoPath,
	})
	if err != nil {
		t.Fatalf("analyze sample repo with configured provider: %v", err)
	}

	if result.AI.Status != "succeeded" || !result.AI.Used {
		t.Fatalf("expected analyze result to surface synthesis success, got %#v", result.AI)
	}
	if result.AI.Provider != "openai-compatible" || result.AI.Model != "gpt-4.1-mini" {
		t.Fatalf("expected analyze result to preserve provider identity, got %#v", result.AI)
	}

	aiResultContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-result.json"))
	if err != nil {
		t.Fatalf("read ai result contract: %v", err)
	}

	var aiResult struct {
		Status                string `json:"status"`
		Used                  bool   `json:"used"`
		ProjectSummary        string `json:"project_summary"`
		ArchitectureNarrative string `json:"architecture_narrative"`
	}
	if err := json.Unmarshal(aiResultContents, &aiResult); err != nil {
		t.Fatalf("unmarshal ai result contract: %v", err)
	}
	if aiResult.Status != "succeeded" || !aiResult.Used {
		t.Fatalf("expected configured provider to produce ai result, got %#v", aiResult)
	}
	if aiResult.ProjectSummary == "" || aiResult.ArchitectureNarrative == "" {
		t.Fatalf("expected synthesized narrative to be persisted, got %#v", aiResult)
	}
}

func TestAnalyzeFallsBackWhenProviderSynthesisFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "not-json"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: repoPath,
	})
	if err != nil {
		t.Fatalf("analyze sample repo with broken provider response: %v", err)
	}

	if result.AI.Status != "fallback" {
		t.Fatalf("expected analyze result to surface fallback status, got %#v", result.AI)
	}
	if !strings.Contains(result.AI.Note, "parse synthesis response") {
		t.Fatalf("expected analyze result to surface parse failure note, got %#v", result.AI)
	}

	aiResultContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "ai-result.json"))
	if err != nil {
		t.Fatalf("read ai result contract: %v", err)
	}

	var aiResult struct {
		Status         string `json:"status"`
		FallbackReason string `json:"fallback_reason"`
	}
	if err := json.Unmarshal(aiResultContents, &aiResult); err != nil {
		t.Fatalf("unmarshal ai result contract: %v", err)
	}
	if aiResult.Status != "fallback" {
		t.Fatalf("expected broken provider response to fall back, got %#v", aiResult)
	}
	if !strings.Contains(aiResult.FallbackReason, "parse synthesis response") {
		t.Fatalf("expected parse failure to be surfaced, got %#v", aiResult)
	}
}
