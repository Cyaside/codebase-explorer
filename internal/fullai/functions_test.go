package fullai

import (
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
)

func TestFunctionPreparerBuildsJobsFromEvidence(t *testing.T) {
	t.Parallel()

	preparer := NewFunctionPreparer()
	functions := preparer.Prepare(Input{
		Analysis: analyzer.Result{
			GeneratedAt: time.Date(2026, time.April, 10, 3, 0, 0, 0, time.UTC),
		},
	}, Plan{
		Mode: string(ModeFull),
	}, Evidence{
		Items: []EvidenceItem{
			{DisplayPath: "cmd/codearch/main.go", Source: "reading-path", Resolution: "exact-file"},
			{DisplayPath: "internal/app/service.go", Source: "module", Resolution: "module-representative"},
			{DisplayPath: "notes/issues.md", Source: "support-file", Resolution: "support-file"},
		},
	})

	if functions.SchemaVersion != FunctionsSchemaVersion {
		t.Fatalf("expected function set schema version %q, got %#v", FunctionsSchemaVersion, functions)
	}
	if len(functions.Jobs) != 7 {
		t.Fatalf("expected all function jobs to be prepared, got %#v", functions.Jobs)
	}
	if functions.Jobs[0].Status != "prepared" {
		t.Fatalf("expected prepared status, got %#v", functions.Jobs[0])
	}
}
