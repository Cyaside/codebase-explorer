package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkbenchOpensLegacyBundleWithoutStructuredGraphs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bundle := filepath.Join(root, "legacy")
	if err := os.MkdirAll(filepath.Join(bundle, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(bundle, "ui"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"data/contract.json": `{ "bundle_schema_version": "phase1.v1" }`,
		"ui/index.html":      `<html></html>`,
		"ui/viewer-data.js":  `window.CODEARCH_VIEWER_DATA = {"bundle_name":"legacy","project":{"name":"Old project"},"metrics":{"total_files":1},"ai":{"status":"succeeded"},"mermaid":{"architecture":"graph TD"}};`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(bundle, filepath.FromSlash(name)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	loaded, err := loadWorkbenchBundle(root, "legacy")
	if err != nil {
		t.Fatalf("open old bundle: %v", err)
	}
	if loaded.Summary.ProjectName != "Old project" || loaded.Data.Mermaid.Architecture != "graph TD" {
		t.Fatalf("legacy bundle data was not retained: %#v", loaded)
	}
	summaries, warnings, err := listWorkbenchBundles(root, 10)
	if err != nil || len(warnings) != 0 || len(summaries) != 1 {
		t.Fatalf("legacy summary fallback failed: summaries=%#v warnings=%#v err=%v", summaries, warnings, err)
	}
}
