package app

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/ui"
)

func TestWorkbenchServesNestedAssetChunks(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	handler, err := service.workbenchHandler(service.settings.DefaultOutputRoot)
	if err != nil {
		t.Fatalf("build workbench handler: %v", err)
	}

	assets, err := ui.Assets()
	if err != nil {
		t.Fatalf("load workbench assets: %v", err)
	}
	entries, err := fs.ReadDir(assets, "chunks")
	if err != nil || len(entries) == 0 {
		t.Fatalf("read embedded chunk assets: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/assets/chunks/"+entries[0].Name(), nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected nested chunk asset to be served, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

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
		BundleWarnings     []string                  `json:"bundle_warnings"`
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
	if len(response.BundleWarnings) != 0 {
		t.Fatalf("expected no bundle warnings, got %#v", response.BundleWarnings)
	}
}

func TestWorkbenchStatusReportsSkippedBundleWarnings(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	bundlePath := filepath.Join(outputRoot, "2026-04-04_010101_broken-bundle")
	if err := os.MkdirAll(filepath.Join(bundlePath, "ui"), 0o755); err != nil {
		t.Fatalf("create broken bundle: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(bundlePath, "data"), 0o755); err != nil {
		t.Fatalf("create broken bundle data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "README.md"), []byte("# broken\n"), 0o644); err != nil {
		t.Fatalf("write broken bundle readme: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "data", "contract.json"), []byte(`{"version":"test"}`), 0o644); err != nil {
		t.Fatalf("write broken bundle contract: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "ui", "index.html"), []byte("<!doctype html><title>broken</title>"), 0o644); err != nil {
		t.Fatalf("write broken viewer index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "ui", "viewer-data.js"), []byte("window.CODEARCH_VIEWER_DATA = invalid;"), 0o644); err != nil {
		t.Fatalf("write broken viewer payload: %v", err)
	}

	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	handler, err := service.workbenchHandler(outputRoot)
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
		RecentBundles  []workbenchBundleSummary `json:"recent_bundles"`
		BundleWarnings []string                 `json:"bundle_warnings"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if len(response.RecentBundles) != 0 {
		t.Fatalf("expected broken bundle to be skipped, got %#v", response.RecentBundles)
	}
	if len(response.BundleWarnings) != 1 {
		t.Fatalf("expected one bundle warning, got %#v", response.BundleWarnings)
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

func TestWorkbenchAnalyzeAcceptsFullAIMode(t *testing.T) {
	t.Parallel()

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
		"repo_path":       filepath.Join("..", "..", "testdata", "sample-repo"),
		"ai_mode":         "full-ai",
		"ai_read_budget":  5,
		"ai_token_budget": 24000,
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
		Result AnalyzeResult `json:"result"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode analyze response: %v", err)
	}
	if !response.Result.FullAI.Enabled || response.Result.FullAI.Status != "prepared" {
		t.Fatalf("expected workbench analyze to carry prepared full-ai summary, got %#v", response.Result.FullAI)
	}
}

func TestWorkbenchBundleDeleteRemovesBundleDirectory(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          filepath.Join("..", "..", "testdata", "sample-repo"),
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("seed bundle: %v", err)
	}

	bundleName := filepath.Base(result.OutputPath)
	handler, err := service.workbenchHandler(service.settings.DefaultOutputRoot)
	if err != nil {
		t.Fatalf("build workbench handler: %v", err)
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/bundles/"+bundleName, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected delete status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if _, err := os.Stat(result.OutputPath); !os.IsNotExist(err) {
		t.Fatalf("expected bundle to be deleted, got stat err %v", err)
	}
}
