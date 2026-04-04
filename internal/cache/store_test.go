package cache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func TestStoreSavesAndLoadsDeterministicPayload(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir(), true)
	payload := DeterministicPayload{
		CachedAt:   time.Date(2026, time.April, 4, 3, 0, 0, 0, time.UTC),
		Key:        "cache-key",
		AppVersion: "test",
		Analysis: analyzer.Result{
			ProjectName: "Codebase Explorer",
		},
	}

	if err := store.SaveDeterministic(payload); err != nil {
		t.Fatalf("save deterministic payload: %v", err)
	}

	loaded, err := store.LoadDeterministic(payload.Key)
	if err != nil {
		t.Fatalf("load deterministic payload: %v", err)
	}
	if loaded.Analysis.ProjectName != "Codebase Explorer" {
		t.Fatalf("expected cached analysis to round-trip, got %#v", loaded)
	}
}

func TestStoreReturnsCacheMissForMissingPayload(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir(), true)
	_, err := store.LoadProvider("missing")
	if !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected cache miss, got %v", err)
	}
}

func TestStoreClearRemovesEntries(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir(), true)
	if err := store.SaveDeterministic(DeterministicPayload{Key: "one"}); err != nil {
		t.Fatalf("seed deterministic cache: %v", err)
	}
	if err := store.SaveProvider(ProviderPayload{Key: "two"}); err != nil {
		t.Fatalf("seed provider cache: %v", err)
	}

	result, err := store.Clear()
	if err != nil {
		t.Fatalf("clear cache: %v", err)
	}
	if result.RemovedEntries == 0 {
		t.Fatalf("expected cache clear to remove entries")
	}

	if _, err := store.LoadDeterministic("one"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected deterministic cache entry to be removed, got %v", err)
	}
	if _, err := store.LoadProvider("two"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected provider cache entry to be removed, got %v", err)
	}
}

func TestBuildDeterministicKeyChangesWhenSupportFilesChange(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := osWriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write repository file: %v", err)
	}

	supportPath := filepath.Join(root, "notes.md")
	if err := osWriteFile(supportPath, []byte("# notes\n"), 0o644); err != nil {
		t.Fatalf("write support file: %v", err)
	}

	firstKey, err := BuildDeterministicKey(root, nil, []string{supportPath}, "test", true)
	if err != nil {
		t.Fatalf("build first key: %v", err)
	}
	if err := osWriteFile(supportPath, []byte("# changed\n"), 0o644); err != nil {
		t.Fatalf("rewrite support file: %v", err)
	}

	secondKey, err := BuildDeterministicKey(root, nil, []string{supportPath}, "test", true)
	if err != nil {
		t.Fatalf("build second key: %v", err)
	}
	if firstKey == secondKey {
		t.Fatalf("expected support file changes to invalidate deterministic cache key")
	}
}

func osWriteFile(path string, contents []byte, mode uint32) error {
	return os.WriteFile(path, contents, os.FileMode(mode))
}
