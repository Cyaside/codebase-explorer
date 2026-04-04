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
	if !strings.Contains(string(indexHTML), "Open a local project") {
		t.Fatalf("expected workbench shell content in index.html")
	}

	appJS, err := fs.ReadFile(assets, "app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	if !strings.Contains(string(appJS), "OpenRouter") || !strings.Contains(string(appJS), "Mistral") {
		t.Fatalf("expected provider presets in app.js")
	}
}
