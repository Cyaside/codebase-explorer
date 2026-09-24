package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestWorkbenchSavedConnectionRunsWithoutResendingKey(t *testing.T) {
	t.Parallel()
	const secret = "fixture-private-api-key"
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var payload struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Messages) < 2 {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		content, _ := json.Marshal(fixtureBatchOutput(payload.Messages[1].Content, "README.md"))
		if strings.Contains(payload.Messages[1].Content, providerDiagnosticsPrompt) {
			content = []byte(secret)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
	}))
	defer providerServer.Close()

	output := t.TempDir()
	service := New(config.Settings{AppVersion: "test", DefaultOutputRoot: output})
	service.credentials = newCredentialStore(filepath.Join(t.TempDir(), "credentials.json"))
	handler, err := service.workbenchHandler(output)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	post := func(path string, payload any) (int, []byte) {
		t.Helper()
		body, _ := json.Marshal(payload)
		response, err := http.Post(server.URL+path, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		contents, _ := io.ReadAll(response.Body)
		return response.StatusCode, contents
	}
	status, body := post("/api/credentials", map[string]string{"id": "fixture", "label": "Fixture", "model": "fixture-model", "base_url": providerServer.URL, "api_key": secret})
	if status != http.StatusOK || bytes.Contains(body, []byte(secret)) {
		t.Fatalf("save connection returned status %d, body %s", status, body)
	}
	status, body = post("/api/credentials", map[string]string{"id": "fixture", "label": "Fixture updated", "model": "fixture-model", "base_url": providerServer.URL})
	if status != http.StatusOK {
		t.Fatalf("replace saved connection returned %d: %s", status, body)
	}
	response, err := http.Get(server.URL + "/api/credentials")
	if err != nil {
		t.Fatal(err)
	}
	listBody, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || !bytes.Contains(listBody, []byte("Fixture updated")) || bytes.Contains(listBody, []byte(secret)) {
		t.Fatalf("connection list leaked key or failed: %d %s", response.StatusCode, listBody)
	}
	doctor, err := service.Doctor(t.Context(), DoctorRequest{CredentialID: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	providerReady := false
	for _, check := range doctor.Checks {
		if check.Name == "provider" && check.Status == "pass" {
			providerReady = true
		}
	}
	if !providerReady {
		t.Fatalf("doctor did not accept saved connection: %#v", doctor.Checks)
	}
	cliResult, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:     filepath.Join("..", "..", "testdata", "sample-repo"),
		CredentialID: "fixture",
	})
	if err != nil || cliResult.OutputPath == "" {
		t.Fatalf("saved connection analyze failed: %v, %#v", err, cliResult)
	}

	providerConfig := map[string]string{"name": "compatible", "model": "fixture-model", "base_url": providerServer.URL}
	status, body = post("/api/provider/test", map[string]any{"provider": providerConfig, "credential_id": "fixture"})
	if status != http.StatusOK || bytes.Contains(body, []byte(secret)) {
		t.Fatalf("saved provider test failed or leaked key: %d %s", status, body)
	}
	status, body = post("/api/analyze", map[string]any{
		"repo_path": filepath.Join("..", "..", "testdata", "sample-repo"),
		"provider":  providerConfig, "credential_id": "fixture",
	})
	if status != http.StatusOK || bytes.Contains(body, []byte(secret)) {
		t.Fatalf("saved connection analysis failed or leaked key: %d %s", status, body)
	}
	var result struct {
		Result struct {
			OutputPath string `json:"output_path"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	viewer, err := os.ReadFile(filepath.Join(result.Result.OutputPath, "ui", "viewer-data.js"))
	if err != nil || bytes.Contains(viewer, []byte(secret)) {
		t.Fatalf("bundle missing or leaked key: %v", err)
	}
	status, body = post("/api/provider/test", map[string]any{
		"provider":      map[string]string{"name": "compatible", "model": "fixture-model", "base_url": "https://different.example/v1"},
		"credential_id": "fixture",
	})
	if status != http.StatusBadRequest || !strings.Contains(string(body), "different base URL") {
		t.Fatalf("saved key was accepted for a different endpoint: %d %s", status, body)
	}
}
