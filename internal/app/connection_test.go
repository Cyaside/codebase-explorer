package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

var testEvidencePath = regexp.MustCompile(`"display_path"\s*:\s*"([^"]+)"`)

func mockConnection(t *testing.T) config.ProviderSettings {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		path := "README.md"
		if len(request.Messages) > 1 {
			if match := testEvidencePath.FindStringSubmatch(request.Messages[1].Content); len(match) > 1 {
				path = match[1]
			}
		}
		content, _ := json.Marshal(fixtureBatchOutput(request.Messages[1].Content, path))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": string(content)}}}})
	}))
	t.Cleanup(server.Close)
	return config.ProviderSettings{Name: "compatible", Model: "fixture-model", APIKey: "fixture-key", BaseURL: server.URL}
}

func TestMediumRepositoryUsesAtMostTwoProviderCalls(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		writeFixtureBatch(w, r)
	}))
	defer server.Close()
	service := New(config.Settings{AppVersion: "test", DefaultOutputRoot: t.TempDir(), Provider: config.ProviderSettings{
		Name: "compatible", Model: "fixture-model", APIKey: "fixture-key", BaseURL: server.URL,
	}})
	result, err := service.Analyze(t.Context(), AnalyzeRequest{RepoPath: filepath.Join("..", "..")})
	if err != nil {
		t.Fatalf("medium repository analysis failed: %v", err)
	}
	if result.TotalFiles < 100 || calls.Load() < 1 || calls.Load() > 2 || result.FullAI.Status != "executed" {
		t.Fatalf("unexpected medium run: files=%d calls=%d status=%s", result.TotalFiles, calls.Load(), result.FullAI.Status)
	}
	t.Logf("medium fixture: files=%d provider_calls=%d prompt_bytes=%d", result.TotalFiles, calls.Load(), result.FullAI.PromptBytes)
}

func TestSecretFixtureNeverEntersProviderPayloadOrBundle(t *testing.T) {
	const secret = "sk-abcdefghijklmnopqrstuvwxyz1234567890"
	root := t.TempDir()
	for name, content := range map[string]string{
		"README.md": "Small Go service with main.go", ".env": "API_KEY=" + secret,
		"main.go": "package main\n// password = " + secret + "\nfunc main() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var leaked atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), secret) || strings.Contains(string(body), ".env") {
			leaked.Store(true)
		}
		var request struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &request)
		content, _ := json.Marshal(fixtureBatchOutput(request.Messages[1].Content, "README.md"))
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
	}))
	defer server.Close()
	service := New(config.Settings{AppVersion: "test", DefaultOutputRoot: t.TempDir(), Provider: config.ProviderSettings{
		Name: "compatible", Model: "fixture", APIKey: "fixture-key", BaseURL: server.URL,
	}})
	result, err := service.Analyze(t.Context(), AnalyzeRequest{RepoPath: root})
	if err != nil {
		t.Fatal(err)
	}
	if leaked.Load() {
		t.Fatal("secret fixture entered provider payload")
	}
	err = filepath.WalkDir(result.OutputPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(contents), secret) || strings.Contains(string(contents), ".env") {
			t.Errorf("secret fixture entered bundle file %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func fixtureBatchOutput(userPrompt string, fallbackPath string) map[string]any {
	var request struct {
		Functions []struct {
			Name                 string   `json:"name"`
			AllowedEvidencePaths []string `json:"allowed_evidence_paths"`
		} `json:"functions"`
	}
	_ = json.Unmarshal([]byte(userPrompt), &request)
	functions := map[string]any{}
	for _, function := range request.Functions {
		path := fallbackPath
		if len(function.AllowedEvidencePaths) > 0 {
			path = function.AllowedEvidencePaths[0]
		}
		if strings.TrimSpace(path) == "" {
			path = "README.md"
		}
		functions[function.Name] = map[string]any{
			"summary":         "Fixture analysis",
			"key_findings":    []any{map[string]any{"claim": "Fixture finding", "evidence_paths": []string{path}, "confidence": "medium"}},
			"recommendations": []string{"Inspect the referenced file."},
			"graph_edges":     []any{map[string]any{"from": path, "to": path, "label": "related", "evidence_paths": []string{path}}},
			"issue_signals":   []any{map[string]any{"title": "Fixture signal", "severity": "low", "evidence_paths": []string{path}}},
		}
	}
	return map[string]any{"functions": functions}
}

func writeFixtureBatch(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Messages) < 2 {
		http.Error(w, "invalid fixture request", http.StatusBadRequest)
		return
	}
	content, _ := json.Marshal(fixtureBatchOutput(request.Messages[1].Content, "README.md"))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
}
