package report

import (
	"strings"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func TestArchitectureMermaidIncludesProjectAndModules(t *testing.T) {
	t.Parallel()

	output := ArchitectureMermaid(testAnalysis())

	if !strings.Contains(output, "flowchart TD") {
		t.Fatalf("expected architecture mermaid header, got %q", output)
	}
	if !strings.Contains(output, "Codebase Explorer<br/>Go CLI application") {
		t.Fatalf("expected project node label, got %q", output)
	}
	if !strings.Contains(output, "Module: internal/app<br/>3 files / 240 lines") {
		t.Fatalf("expected module node label, got %q", output)
	}
	if !strings.Contains(output, "Hotspot: internal/app/service.go<br/>score 8.20") {
		t.Fatalf("expected hotspot node label, got %q", output)
	}
}

func TestDependenciesMermaidIncludesRiskBuckets(t *testing.T) {
	t.Parallel()

	analysis := testAnalysis()
	analysis.DependencyRisks = []analyzer.DependencyRisk{
		{
			Path:        "internal/app/service.go",
			ImportCount: 7,
			Reason:      "high import concentration",
		},
	}

	output := DependenciesMermaid(analysis)

	if !strings.Contains(output, "flowchart LR") {
		t.Fatalf("expected dependencies mermaid header, got %q", output)
	}
	if !strings.Contains(output, "Dependency concentration review") {
		t.Fatalf("expected dependency root label, got %q", output)
	}
	if !strings.Contains(output, "Module: internal") {
		t.Fatalf("expected module bucket label, got %q", output)
	}
	if !strings.Contains(output, "internal/app/service.go<br/>imports 7") {
		t.Fatalf("expected dependency risk node label, got %q", output)
	}
}
