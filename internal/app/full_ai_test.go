package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
)

func TestAnalyzeWritesFullAIPlanWhenRequested(t *testing.T) {
	t.Parallel()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
	})

	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: filepath.Join("..", "..", "testdata", "sample-repo"),
		FullAI: fullai.Options{
			Mode:       fullai.ModeFull,
			ReadBudget: 6,
		},
	})
	if err != nil {
		t.Fatalf("analyze sample repo with full-ai scaffold: %v", err)
	}

	if !result.FullAI.Enabled || result.FullAI.Status != "planned" {
		t.Fatalf("expected planned full-ai summary, got %#v", result.FullAI)
	}

	planContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "full-ai-plan.json"))
	if err != nil {
		t.Fatalf("read full-ai plan: %v", err)
	}

	var plan struct {
		SchemaVersion string `json:"schema_version"`
		ReadBudget    int    `json:"read_budget"`
	}
	if err := json.Unmarshal(planContents, &plan); err != nil {
		t.Fatalf("unmarshal full-ai plan: %v", err)
	}
	if plan.SchemaVersion != fullai.PlanSchemaVersion {
		t.Fatalf("expected full-ai plan schema version %q, got %#v", fullai.PlanSchemaVersion, plan)
	}
	if plan.ReadBudget != 6 {
		t.Fatalf("expected configured full-ai read budget, got %#v", plan)
	}
}
