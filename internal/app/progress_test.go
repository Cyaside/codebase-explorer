package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeEmitsProgressForDeterministicOnly(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	var events []AnalyzeProgressEvent
	_, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          filepath.Join("..", "..", "testdata", "sample-repo"),
		DeterministicOnly: true,
		Progress: func(event AnalyzeProgressEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("analyze sample repo: %v", err)
	}

	if len(events) < 2 {
		t.Fatalf("expected AI progress events to be emitted, got %#v", events)
	}
	if events[0].Stage != "ai-context" || events[0].Status != "ready" {
		t.Fatalf("expected first progress event to describe AI context, got %#v", events)
	}
	if events[1].Stage != "ai-synthesis" || events[1].Status != "skipped" {
		t.Fatalf("expected second progress event to describe skipped synthesis, got %#v", events)
	}
}

func TestAnalyzeEmitsProgressForSuccessfulSynthesis(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"project_summary\":\"AI summary\"}"
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

	var events []AnalyzeProgressEvent
	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: filepath.Join("..", "..", "testdata", "sample-repo"),
		Progress: func(event AnalyzeProgressEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("analyze sample repo: %v", err)
	}

	if result.AI.Status != "succeeded" {
		t.Fatalf("expected AI summary to capture success, got %#v", result.AI)
	}
	if len(events) < 3 {
		t.Fatalf("expected progress events for AI context and synthesis, got %#v", events)
	}
	if events[1].Status != "running" {
		t.Fatalf("expected running synthesis event, got %#v", events)
	}
	if events[2].Status != "succeeded" {
		t.Fatalf("expected successful synthesis event, got %#v", events)
	}
}
