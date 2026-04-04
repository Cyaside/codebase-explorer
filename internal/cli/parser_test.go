package cli

import "testing"

func TestParseAnalyzeAcceptsExplicitSupportFlags(t *testing.T) {
	t.Parallel()

	command, err := parse([]string{
		"analyze",
		"./repo",
		"--support", "notes.md",
		"--issues", "issues.json",
		"--changelog", "CHANGELOG.md",
	})
	if err != nil {
		t.Fatalf("parse analyze command: %v", err)
	}

	if len(command.analyzeRequest.OptionalSupportFiles) != 3 {
		t.Fatalf("expected three support files, got %#v", command.analyzeRequest.OptionalSupportFiles)
	}
}

func TestParseAnalyzeRejectsMissingSupportValue(t *testing.T) {
	t.Parallel()

	_, err := parse([]string{"analyze", "./repo", "--support"})
	if err == nil {
		t.Fatalf("expected parse error for missing support value")
	}
}
