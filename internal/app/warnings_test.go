package app

import (
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func TestBuildWarningsFlagsLargeInputs(t *testing.T) {
	t.Parallel()

	warnings := buildWarnings(analyzer.Result{
		Metrics: analyzer.Metrics{
			TotalFiles: 6000,
			TotalLines: 600000,
		},
	}, 12)

	if len(warnings) != 3 {
		t.Fatalf("expected three warnings, got %#v", warnings)
	}
	if !strings.Contains(warnings[0], "Large repository detected") {
		t.Fatalf("expected repo-size warning, got %#v", warnings)
	}
}
