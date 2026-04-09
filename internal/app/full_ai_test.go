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

	if !result.FullAI.Enabled || result.FullAI.Status != "prepared" {
		t.Fatalf("expected prepared full-ai summary, got %#v", result.FullAI)
	}
	if result.FullAI.CollectedItems == 0 {
		t.Fatalf("expected collected evidence items in full-ai summary, got %#v", result.FullAI)
	}
	if result.FullAI.PreparedFunctions == 0 {
		t.Fatalf("expected prepared full-ai function jobs in summary, got %#v", result.FullAI)
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

	evidenceContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "full-ai-evidence.json"))
	if err != nil {
		t.Fatalf("read full-ai evidence: %v", err)
	}
	var evidence struct {
		SchemaVersion  string `json:"schema_version"`
		CollectedItems int    `json:"collected_items"`
	}
	if err := json.Unmarshal(evidenceContents, &evidence); err != nil {
		t.Fatalf("unmarshal full-ai evidence: %v", err)
	}
	if evidence.SchemaVersion != fullai.EvidenceSchemaVersion {
		t.Fatalf("expected full-ai evidence schema version %q, got %#v", fullai.EvidenceSchemaVersion, evidence)
	}
	if evidence.CollectedItems == 0 {
		t.Fatalf("expected collected evidence items, got %#v", evidence)
	}

	functionContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "full-ai-functions.json"))
	if err != nil {
		t.Fatalf("read full-ai functions: %v", err)
	}
	var functionSet struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal(functionContents, &functionSet); err != nil {
		t.Fatalf("unmarshal full-ai functions: %v", err)
	}
	if functionSet.SchemaVersion != fullai.FunctionsSchemaVersion {
		t.Fatalf("expected full-ai functions schema version %q, got %#v", fullai.FunctionsSchemaVersion, functionSet)
	}
}
