package repo

import "testing"

func TestMatcherIgnoresDefaultDirectories(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher(nil)
	if !matcher.ShouldIgnore("node_modules", true) {
		t.Fatalf("expected node_modules directory to be ignored")
	}
}

func TestMatcherIgnoresPatternsFromFiles(t *testing.T) {
	t.Parallel()

	matcher := NewMatcher([]string{"build/", "*.log", "docs/*.md"})
	if !matcher.ShouldIgnore("build", true) {
		t.Fatalf("expected build directory to be ignored")
	}
	if !matcher.ShouldIgnore("server.log", false) {
		t.Fatalf("expected log file to be ignored")
	}
	if !matcher.ShouldIgnore("docs/readme.md", false) {
		t.Fatalf("expected docs/readme.md to be ignored")
	}
	if matcher.ShouldIgnore("cmd/main.go", false) {
		t.Fatalf("did not expect cmd/main.go to be ignored")
	}
}
