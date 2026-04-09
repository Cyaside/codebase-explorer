package fullai

import (
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/changes"
)

func TestPlannerBuildsPlanForFullAIMode(t *testing.T) {
	t.Parallel()

	planner := NewPlanner()
	plan, summary := planner.Plan(Input{
		RootPath: "sample-repo",
		Analysis: analyzer.Result{
			GeneratedAt: time.Date(2026, time.April, 10, 1, 0, 0, 0, time.UTC),
			ReadingPath: []analyzer.ReadingPathItem{
				{Path: "cmd/codearch/main.go", Reason: "entry point"},
				{Path: "internal/app/service.go", Reason: "application orchestration"},
			},
			Hotspots: []analyzer.Hotspot{
				{Path: "internal/app/service.go", Reasons: []string{"many responsibilities"}},
			},
			DependencyRisks: []analyzer.DependencyRisk{
				{Path: "internal/provider/openai_compatible.go", Reason: "external provider boundary"},
			},
			Modules: []analyzer.ModuleInfo{
				{Path: "internal/app", FileCount: 6, TotalLines: 700},
			},
		},
		Changes: changes.Result{
			FrequentlyMentionedAreas: []changes.AreaMention{
				{Path: "internal/app/service.go", Reasons: []string{"touched in multiple issue notes"}},
			},
		},
		SupportFiles: []string{"notes/issues.md"},
	}, Options{
		Mode:       ModeFull,
		ReadBudget: 4,
	})
	if !summary.Enabled {
		t.Fatalf("expected full-ai summary to be enabled, got %#v", summary)
	}
	if summary.Status != "planned" {
		t.Fatalf("expected planned summary status, got %#v", summary)
	}
	if len(plan.Targets) != 4 {
		t.Fatalf("expected plan to respect read budget, got %#v", plan.Targets)
	}
	if !plan.Truncated {
		t.Fatalf("expected plan to record truncation when targets exceed budget")
	}
}

func TestPlannerReturnsDisabledSummaryForStandardMode(t *testing.T) {
	t.Parallel()

	planner := NewPlanner()
	plan, summary := planner.Plan(Input{
		RootPath: "sample-repo",
		Analysis: analyzer.Result{
			GeneratedAt: time.Date(2026, time.April, 10, 1, 0, 0, 0, time.UTC),
		},
	}, Options{})
	if summary.Enabled {
		t.Fatalf("expected disabled summary, got %#v", summary)
	}
	if summary.Status != "disabled" {
		t.Fatalf("expected disabled summary status, got %#v", summary)
	}
	if plan.SchemaVersion != PlanSchemaVersion {
		t.Fatalf("expected plan schema version %q, got %#v", PlanSchemaVersion, plan)
	}
}
