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

func TestParseServeAcceptsAddrAndNoBrowser(t *testing.T) {
	t.Parallel()

	command, err := parse([]string{"serve", "--addr", "127.0.0.1:4180", "--no-browser"})
	if err != nil {
		t.Fatalf("parse serve command: %v", err)
	}
	if command.name != "serve" {
		t.Fatalf("expected serve command, got %#v", command)
	}
	if command.serveRequest.Addr != "127.0.0.1:4180" || !command.serveRequest.NoBrowser {
		t.Fatalf("expected parsed serve request, got %#v", command.serveRequest)
	}
}

func TestParseStartAliasAcceptsServeFlags(t *testing.T) {
	t.Parallel()

	command, err := parse([]string{"start", "--addr", "127.0.0.1:4215", "--no-browser"})
	if err != nil {
		t.Fatalf("parse start command: %v", err)
	}
	if command.name != "serve" {
		t.Fatalf("expected start alias to map to serve command, got %#v", command)
	}
	if command.serveRequest.Addr != "127.0.0.1:4215" || !command.serveRequest.NoBrowser {
		t.Fatalf("expected parsed serve request from start alias, got %#v", command.serveRequest)
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
