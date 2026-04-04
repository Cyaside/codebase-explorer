package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
)

func TestOpenResolvesLatestBundleWhenPathMissing(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	oldBundle := createOpenBundle(t, outputRoot, "2026-04-04_090000_old")
	latestBundle := createOpenBundle(t, outputRoot, "2026-04-04_100000_latest")

	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		AppVersion:        "test",
		ConfigSource:      "test",
		OutputKeepLatest:  10,
	})

	result, err := service.Open(t.Context(), OpenRequest{})
	if err != nil {
		t.Fatalf("open latest bundle: %v", err)
	}

	if result.BundlePath != latestBundle {
		t.Fatalf("expected latest bundle to be selected, got %#v", result)
	}
	if !result.ResolvedLatest {
		t.Fatalf("expected result to note latest bundle resolution")
	}
	if _, err := os.Stat(oldBundle); err != nil {
		t.Fatalf("expected older bundle to remain untouched: %v", err)
	}
}

func TestOpenAcceptsExplicitBundlePath(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	explicitBundle := createOpenBundle(t, outputRoot, "2026-04-04_100000_bundle")

	service := New(config.Settings{
		DefaultOutputRoot: outputRoot,
		AppVersion:        "test",
		ConfigSource:      "test",
		OutputKeepLatest:  10,
	})

	result, err := service.Open(t.Context(), OpenRequest{BundlePath: explicitBundle})
	if err != nil {
		t.Fatalf("open explicit bundle: %v", err)
	}

	if result.BundlePath != explicitBundle {
		t.Fatalf("expected explicit bundle path to be used, got %#v", result)
	}
	if result.ResolvedLatest {
		t.Fatalf("did not expect explicit bundle to be marked as latest-resolved")
	}
}

func createOpenBundle(t *testing.T, outputRoot, name string) string {
	t.Helper()

	bundlePath := filepath.Join(outputRoot, name)
	if err := os.MkdirAll(filepath.Join(bundlePath, "data"), 0o755); err != nil {
		t.Fatalf("create bundle data directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(bundlePath, "ui"), 0o755); err != nil {
		t.Fatalf("create bundle ui directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "data", "contract.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write contract file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "ui", "index.html"), []byte("<html></html>\n"), 0o644); err != nil {
		t.Fatalf("write viewer index: %v", err)
	}

	return bundlePath
}
