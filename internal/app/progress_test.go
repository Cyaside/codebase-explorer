package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
)

func TestAnalyzeEmitsProgressForSingleWorkflow(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider:          mockConnection(t),
	})

	var events []AnalyzeProgressEvent
	_, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: filepath.Join("..", "..", "testdata", "sample-repo"),
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
	if !hasProgressEvent(events, "ai-context", "ready") {
		t.Fatalf("expected AI context progress event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-execution", "executed") {
		t.Fatalf("expected completed AI execution event, got %#v", events)
	}
}

func TestAnalyzeEmitsProgressForSuccessfulSynthesis(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeFixtureBatch(w, r)
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
	if !hasProgressEvent(events, "full-ai-execution", "running") {
		t.Fatalf("expected running AI execution event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-execution", "executed") {
		t.Fatalf("expected completed AI execution event, got %#v", events)
	}
}

func TestAnalyzeEmitsProgressForFullAIPlanning(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider:          mockConnection(t),
	})

	var events []AnalyzeProgressEvent
	_, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: filepath.Join("..", "..", "testdata", "sample-repo"),
		FullAI:   fullai.Options{Mode: fullai.ModeFull},
		Progress: func(event AnalyzeProgressEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("analyze sample repo with full-ai scaffold: %v", err)
	}

	if !hasProgressEvent(events, "full-ai-plan", "running") {
		t.Fatalf("expected running full-ai planning event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-plan", "planned") {
		t.Fatalf("expected planned full-ai planning event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-evidence", "running") {
		t.Fatalf("expected running full-ai evidence event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-evidence", "collected") {
		t.Fatalf("expected collected full-ai evidence event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-functions", "running") {
		t.Fatalf("expected running full-ai function preparation event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-functions", "prepared") {
		t.Fatalf("expected prepared full-ai function event, got %#v", events)
	}
	if !hasProgressEvent(events, "full-ai-execution", "executed") {
		t.Fatalf("expected completed AI execution event, got %#v", events)
	}
}

func hasProgressEvent(events []AnalyzeProgressEvent, stage string, status string) bool {
	for _, event := range events {
		if event.Stage == stage && event.Status == status {
			return true
		}
	}
	return false
}
