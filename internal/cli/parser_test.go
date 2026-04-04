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

func TestParseCacheClearCommand(t *testing.T) {
	t.Parallel()

	command, err := parse([]string{"cache", "clear"})
	if err != nil {
		t.Fatalf("parse cache clear command: %v", err)
	}
	if command.name != "cache-clear" {
		t.Fatalf("expected cache-clear command, got %#v", command)
	}
}

func TestParseExportAcceptsOutputFlag(t *testing.T) {
	t.Parallel()

	command, err := parse([]string{"export", "./bundle", "--output", "./bundle.zip"})
	if err != nil {
		t.Fatalf("parse export command: %v", err)
	}
	if command.name != "export" || command.exportRequest.OutputPath != "./bundle.zip" {
		t.Fatalf("expected parsed export request, got %#v", command)
	}
}
