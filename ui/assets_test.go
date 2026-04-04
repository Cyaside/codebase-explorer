package ui

import (
	"io/fs"
	"strings"
	"testing"
)

func TestAssetsExposeWorkbenchShell(t *testing.T) {
	t.Parallel()

	assets, err := Assets()
	if err != nil {
		t.Fatalf("assets fs: %v", err)
	}

	indexHTML, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if !strings.Contains(string(indexHTML), "/assets/app.js") {
		t.Fatalf("expected built workbench script reference in index.html")
	}

	appJS, err := fs.ReadFile(assets, "app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	if !strings.Contains(string(appJS), "Codebase Explorer Workbench") {
		t.Fatalf("expected workbench title in bundled app.js")
	}
}
