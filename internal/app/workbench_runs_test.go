package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestWorkbenchAnalyzeRunCompletes(t *testing.T) {
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

	body, err := json.Marshal(map[string]any{
		"repo_path":          filepath.Join("..", "..", "testdata", "sample-repo"),
		"deterministic_only": true,
	})
	if err != nil {
		t.Fatalf("marshal run body: %v", err)
	}

	startRequest := httptest.NewRequest(http.MethodPost, "/api/analyze-runs", bytes.NewReader(body))
	startRecorder := httptest.NewRecorder()
	handler.ServeHTTP(startRecorder, startRequest)

	if startRecorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", startRecorder.Code, startRecorder.Body.String())
	}

	var started struct {
		Run workbenchAnalyzeRun `json:"run"`
	}
	if err := json.Unmarshal(startRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode start response: %v", err)
	}
	if started.Run.ID == "" {
		t.Fatalf("expected run id, got %#v", started.Run)
	}

	run := pollWorkbenchRun(t, handler, started.Run.ID, 50*time.Millisecond, 120)
	if run.Status != "succeeded" {
		t.Fatalf("expected run to succeed, got %#v", run)
	}
	if run.Response == nil {
		t.Fatalf("expected run response to be available, got %#v", run)
	}
	if len(run.Progress) == 0 {
		t.Fatalf("expected progress events, got %#v", run)
	}
}

func TestWorkbenchAnalyzeRunCanBeCanceled(t *testing.T) {
	t.Parallel()

	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(3 * time.Second):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"project_summary\":\"done\",\"architecture_narrative\":\"done\"}"}}]}`))
		}
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

	body, err := json.Marshal(map[string]any{
		"repo_path": filepath.Join("..", "..", "testdata", "sample-repo"),
		"provider": map[string]any{
			"name":     "openai-compatible",
			"model":    "mistral-small-latest",
			"api_key":  "test-key",
			"base_url": providerServer.URL,
		},
	})
	if err != nil {
		t.Fatalf("marshal run body: %v", err)
	}

	startRequest := httptest.NewRequest(http.MethodPost, "/api/analyze-runs", bytes.NewReader(body))
	startRecorder := httptest.NewRecorder()
	handler.ServeHTTP(startRecorder, startRequest)

	if startRecorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", startRecorder.Code, startRecorder.Body.String())
	}

	var started struct {
		Run workbenchAnalyzeRun `json:"run"`
	}
	if err := json.Unmarshal(startRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode start response: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	cancelRequest := httptest.NewRequest(http.MethodPost, "/api/analyze-runs/"+started.Run.ID+"/cancel", nil)
	cancelRecorder := httptest.NewRecorder()
	handler.ServeHTTP(cancelRecorder, cancelRequest)

	if cancelRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", cancelRecorder.Code, cancelRecorder.Body.String())
	}

	run := pollWorkbenchRun(t, handler, started.Run.ID, 100*time.Millisecond, 80)
	if run.Status != "canceled" {
		t.Fatalf("expected run to be canceled, got %#v", run)
	}
	if run.Error == "" {
		t.Fatalf("expected canceled run to carry an error detail, got %#v", run)
	}
}

func pollWorkbenchRun(t *testing.T, handler http.Handler, runID string, wait time.Duration, attempts int) workbenchAnalyzeRun {
	t.Helper()

	for range attempts {
		request := httptest.NewRequest(http.MethodGet, "/api/analyze-runs/"+runID, nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200 while polling run, got %d: %s", recorder.Code, recorder.Body.String())
		}

		var response struct {
			Run workbenchAnalyzeRun `json:"run"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode polled run response: %v", err)
		}

		switch response.Run.Status {
		case "succeeded", "failed", "canceled":
			return response.Run
		}

		time.Sleep(wait)
	}

	t.Fatalf("run %s did not reach a terminal status", runID)
	return workbenchAnalyzeRun{}
}
