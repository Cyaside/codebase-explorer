package bundle

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNextBundleNameAddsSuffixWhenBaseExists(t *testing.T) {
	t.Parallel()

	outputRoot := t.TempDir()
	baseName := "2026-04-04_090000_codebase-explorer"
	if err := os.MkdirAll(filepath.Join(outputRoot, baseName), 0o755); err != nil {
		t.Fatalf("seed existing bundle: %v", err)
	}

	name, err := nextBundleName(outputRoot, "Codebase Explorer", time.Date(2026, time.April, 4, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("allocate bundle name: %v", err)
	}
	if name != baseName+"_001" {
		t.Fatalf("expected suffixed bundle name, got %q", name)
	}
}
