package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestWorkbenchStatusListsRecentBundles(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	if _, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	}); err != nil {
		t.Fatalf("seed bundle: %v", err)
	}

	handler, err := service.workbenchHandler(service.settings.DefaultOutputRoot)
	if err != nil {
		t.Fatalf("build workbench handler: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RecentBundles      []workbenchBundleSummary  `json:"recent_bundles"`
		SupportedProviders []workbenchProviderOption `json:"supported_providers"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if len(response.RecentBundles) != 1 {
		t.Fatalf("expected one recent bundle, got %#v", response.RecentBundles)
	}
	if len(response.SupportedProviders) == 0 {
		t.Fatalf("expected supported providers to be exposed")
	}
}

func TestWorkbenchAnalyzeUsesProviderOverride(t *testing.T) {
	t.Parallel()

	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"project_summary\":\"Workbench summary\",\"architecture_narrative\":\"Workbench architecture\"}"
					}
				}
			]
		}`))
	}))
	defer providerServer.Close()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	handler, err := service.workbenchHandler(service.settings.DefaultOutputRoot)
	if err != nil {
		t.Fatalf("build workbench handler: %v", err)
	}

	requestBody := map[string]any{
		"repo_path": filepath.Join("..", "..", "testdata", "sample-repo"),
		"provider": map[string]any{
			"name":     "openai-compatible",
			"model":    "mistral-small-latest",
			"api_key":  "ui-key",
			"base_url": providerServer.URL,
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Result AnalyzeResult          `json:"result"`
		Bundle workbenchBundleSummary `json:"bundle"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode analyze response: %v", err)
	}
	if response.Result.AI.Status != "succeeded" || !response.Result.AI.Used {
		t.Fatalf("expected workbench analyze to succeed, got %#v", response.Result.AI)
	}
	if response.Bundle.ProjectName == "" {
		t.Fatalf("expected bundle summary to be present, got %#v", response.Bundle)
	}
}
