package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestWorkbenchAnalyzeRunEventsStream(t *testing.T) {
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

	streamRequest := httptest.NewRequest(http.MethodGet, "/api/analyze-runs/"+started.Run.ID+"/events", nil)
	streamRecorder := httptest.NewRecorder()
	handler.ServeHTTP(streamRecorder, streamRequest)

	if streamRecorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", streamRecorder.Code, streamRecorder.Body.String())
	}

	if contentType := streamRecorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected SSE content type, got %q", contentType)
	}

	bodyText := streamRecorder.Body.String()
	if !strings.Contains(bodyText, "event: run") {
		t.Fatalf("expected SSE run events, got %q", bodyText)
	}
	if !strings.Contains(bodyText, "\"status\":\"succeeded\"") {
		t.Fatalf("expected final succeeded snapshot in stream, got %q", bodyText)
	}
}
