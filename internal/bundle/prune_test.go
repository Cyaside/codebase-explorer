package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPruneBundlesKeepsLatestManagedBundles(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	oldestBundle := createManagedBundle(t, outputRoot, "2026-04-04_100001_oldest")
	createManagedBundle(t, outputRoot, "2026-04-04_100002_previous")
	currentBundle := createManagedBundle(t, outputRoot, "2026-04-04_100003_current")
	createUnmanagedDirectory(t, outputRoot, "notes")

	prunedBundles, err := pruneBundles(outputRoot, currentBundle, 2)
	if err != nil {
		t.Fatalf("prune bundles: %v", err)
	}
	if prunedBundles != 1 {
		t.Fatalf("expected one old bundle to be pruned, got %d", prunedBundles)
	}

	if _, err := os.Stat(currentBundle); err != nil {
		t.Fatalf("expected current bundle to remain: %v", err)
	}
	if _, err := os.Stat(oldestBundle); !os.IsNotExist(err) {
		t.Fatalf("expected oldest managed bundle to be removed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputRoot, "notes")); err != nil {
		t.Fatalf("expected unmanaged directory to remain: %v", err)
	}
}

func TestPruneBundlesDisabledWhenRetentionIsZero(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	currentBundle := createManagedBundle(t, outputRoot, "2026-04-04_100002_current")
	olderBundle := createManagedBundle(t, outputRoot, "2026-04-04_100001_old")

	prunedBundles, err := pruneBundles(outputRoot, currentBundle, 0)
	if err != nil {
		t.Fatalf("prune bundles: %v", err)
	}
	if prunedBundles != 0 {
		t.Fatalf("expected pruning to be disabled, got %d", prunedBundles)
	}
	if _, err := os.Stat(olderBundle); err != nil {
		t.Fatalf("expected older bundle to remain when cleanup disabled: %v", err)
	}
}

func createManagedBundle(t *testing.T, outputRoot, name string) string {
	t.Helper()

	bundlePath := filepath.Join(outputRoot, name)
	if err := os.MkdirAll(filepath.Join(bundlePath, "data"), 0o755); err != nil {
		t.Fatalf("create bundle directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "README.md"), []byte("# bundle\n"), 0o644); err != nil {
		t.Fatalf("write bundle readme: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundlePath, "data", "contract.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	return bundlePath
}

func createUnmanagedDirectory(t *testing.T, outputRoot, name string) {
	t.Helper()

	directoryPath := filepath.Join(outputRoot, name)
	if err := os.MkdirAll(directoryPath, 0o755); err != nil {
		t.Fatalf("create unmanaged directory: %v", err)
	}
}
