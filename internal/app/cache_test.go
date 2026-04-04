package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestAnalyzeReusesDeterministicCache(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	cacheRoot := filepath.Join(outputRoot, "cache")
	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		CacheRoot:         cacheRoot,
		CacheEnabled:      true,
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	first, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("first analyze: %v", err)
	}
	if first.Cache.DeterministicStatus != "miss" {
		t.Fatalf("expected first analyze to miss deterministic cache, got %#v", first.Cache)
	}

	second, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	})
	if err != nil {
		t.Fatalf("second analyze: %v", err)
	}
	if second.Cache.DeterministicStatus != "hit" {
		t.Fatalf("expected second analyze to hit deterministic cache, got %#v", second.Cache)
	}
}

func TestAnalyzeReusesProviderCache(t *testing.T) {
	t.Parallel()

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
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

	outputRoot := t.TempDir()
	cacheRoot := filepath.Join(outputRoot, "cache")
	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		CacheRoot:         cacheRoot,
		CacheEnabled:      true,
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
	first, err := service.Analyze(t.Context(), AnalyzeRequest{RepoPath: repoPath})
	if err != nil {
		t.Fatalf("first analyze with provider: %v", err)
	}
	if first.Cache.ProviderStatus != "miss" {
		t.Fatalf("expected first analyze to miss provider cache, got %#v", first.Cache)
	}

	second, err := service.Analyze(t.Context(), AnalyzeRequest{RepoPath: repoPath})
	if err != nil {
		t.Fatalf("second analyze with provider: %v", err)
	}
	if second.Cache.ProviderStatus != "hit" {
		t.Fatalf("expected second analyze to hit provider cache, got %#v", second.Cache)
	}
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Fatalf("expected provider server to be called once, got %d", requestCount)
	}
}

func TestClearCacheRemovesStoredEntries(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	cacheRoot := filepath.Join(outputRoot, "cache")
	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		CacheRoot:         cacheRoot,
		CacheEnabled:      true,
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	repoPath := filepath.Join("..", "..", "testdata", "sample-repo")
	if _, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath:          repoPath,
		DeterministicOnly: true,
	}); err != nil {
		t.Fatalf("seed cache through analyze: %v", err)
	}

	result, err := service.ClearCache(t.Context(), CacheClearRequest{})
	if err != nil {
		t.Fatalf("clear cache: %v", err)
	}
	if result.RemovedEntries == 0 {
		t.Fatalf("expected cache clear to remove at least one entry")
	}
}
