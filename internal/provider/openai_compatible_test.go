package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleClientSynthesize(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("expected chat completions path, got %q", r.URL.Path)
		}
		if authorization := r.Header.Get("Authorization"); authorization != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", authorization)
		}

		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload.Model != "gpt-4.1-mini" {
			t.Fatalf("expected model to be forwarded, got %#v", payload)
		}
		if len(payload.Messages) != 2 {
			t.Fatalf("expected two prompt messages, got %#v", payload)
		}
		if payload.Messages[0].Role != "system" || payload.Messages[1].Role != "user" {
			t.Fatalf("expected system and user messages, got %#v", payload.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"project_summary\":\"CLI analyzer for repository orientation.\",\"architecture_narrative\":\"CLI delegates to app orchestration, which runs repo scanning, deterministic analysis, and bundle writing.\",\"hotspot_explanations\":[{\"path\":\"internal/app/service.go\",\"explanation\":\"This file orchestrates the end-to-end analyze flow.\"}],\"reading_path_explanations\":[{\"path\":\"cmd/codearch/main.go\",\"rationale\":\"Start here to see process bootstrap and CLI wiring.\"}]}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.Client())
	result, err := client.Synthesize(t.Context(), Request{
		Config: Config{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
		Context: testCondensedContext(),
	})
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}

	if !result.Successful() {
		t.Fatalf("expected successful synthesis result, got %#v", result)
	}
	if result.Provider != "openai-compatible" {
		t.Fatalf("expected provider identity to be preserved, got %#v", result)
	}
	if len(result.HotspotExplanations) != 1 {
		t.Fatalf("expected hotspot explanations to be parsed, got %#v", result)
	}
	if len(result.ReadingPathExplanations) != 1 {
		t.Fatalf("expected reading path explanations to be parsed, got %#v", result)
	}
}

func TestOpenAICompatibleClientRejectsMalformedPayload(t *testing.T) {
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

	client := NewOpenAICompatibleClient(server.Client())
	_, err := client.Synthesize(t.Context(), Request{
		Config: Config{
			Name:    "openai-compatible",
			Model:   "gpt-4.1-mini",
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
		Context: testCondensedContext(),
	})
	if err == nil {
		t.Fatalf("expected malformed payload to fail")
	}
}

func TestOpenAICompatibleClientCompleteReturnsRawContent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"summary\":\"Function output\"}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.Client())
	result, err := client.Complete(t.Context(), PromptRequest{
		Config: Config{
			Name:    "openai-compatible",
			Model:   "mistral-small-latest",
			APIKey:  "test-key",
			BaseURL: server.URL,
		},
		SystemPrompt: "system",
		UserPrompt:   "user",
	})
	if err != nil {
		t.Fatalf("complete prompt: %v", err)
	}
	if !result.Used || result.Content == "" {
		t.Fatalf("expected raw prompt content, got %#v", result)
	}
}

func TestParseSynthesisResultStripsMarkdownFence(t *testing.T) {
	t.Parallel()

	result, err := parseSynthesisResult("```json\n{\"project_summary\":\"summary\"}\n```")
	if err != nil {
		t.Fatalf("parse synthesis result: %v", err)
	}
	if result.ProjectSummary != "summary" {
		t.Fatalf("expected fenced JSON to be parsed, got %#v", result)
	}
}

func TestChatCompletionsURLUsesOpenAIDefault(t *testing.T) {
	t.Parallel()

	url := chatCompletionsURL(Config{Name: "openai"})
	if !strings.HasPrefix(url, defaultOpenAIBaseURL) {
		t.Fatalf("expected openai provider to use default base URL, got %q", url)
	}
}

func testCondensedContext() CondensedContext {
	return CondensedContext{
		SchemaVersion: "context.v1",
		Project: ProjectFacts{
			Name:            "codebase-explorer",
			Type:            "Go CLI application",
			Summary:         "Repository orientation tool",
			PrimaryLanguage: "Go",
			TotalFiles:      12,
			TotalLines:      1200,
		},
		Modules: []ModuleSummary{
			{
				Path:            "internal/app",
				FileCount:       3,
				TotalLines:      300,
				Languages:       []string{"Go"},
				EntryPointCount: 0,
				MarkerCount:     0,
			},
		},
		EntryPoints: []string{"cmd/codearch/main.go"},
		Hotspots: []HotspotSummary{
			{
				Path:    "internal/app/service.go",
				Score:   7.5,
				Reasons: []string{"many imports"},
			},
		},
		ReadingPath: []ReadingPathHint{
			{
				Path:   "cmd/codearch/main.go",
				Reason: "CLI bootstrap",
			},
		},
	}
}
