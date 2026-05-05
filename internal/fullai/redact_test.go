package fullai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/repo"
)

func TestEvidenceExcludesSecretFilesAndRedactsOrdinarySource(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	secret := "sk-abcdefghijklmnopqrstuvwxyz123456"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("API_KEY="+secret), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n// token: "+secret+"\npassword = hunter2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	input := Input{RootPath: root, ScanResult: repo.ScanResult{Files: []repo.FileInfo{{Path: "main.go"}, {Path: ".env"}}}}
	skipped := collectEvidenceItem(input, Target{Path: ".env"})
	if skipped.ReadStatus != "sensitive-skipped" || skipped.Snippet != "" {
		t.Fatalf("secret file should be skipped: %#v", skipped)
	}
	item := collectEvidenceItem(input, Target{Path: "main.go"})
	if item.ReadStatus != "read" || strings.Contains(item.Snippet, secret) || strings.Contains(item.Snippet, "hunter2") {
		t.Fatalf("source secrets should be redacted: %#v", item)
	}
}
