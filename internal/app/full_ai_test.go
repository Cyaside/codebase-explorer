package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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

	if !result.FullAI.Enabled || result.FullAI.Status != "provider-disabled" {
		t.Fatalf("expected provider-disabled full-ai summary, got %#v", result.FullAI)
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

	executionContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "full-ai-execution.json"))
	if err != nil {
		t.Fatalf("read full-ai execution: %v", err)
	}
	var execution struct {
		SchemaVersion string `json:"schema_version"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(executionContents, &execution); err != nil {
		t.Fatalf("unmarshal full-ai execution: %v", err)
	}
	if execution.SchemaVersion != fullai.ExecutionSchemaVersion {
		t.Fatalf("expected full-ai execution schema version %q, got %#v", fullai.ExecutionSchemaVersion, execution)
	}

	verificationContents, err := os.ReadFile(filepath.Join(result.OutputPath, "data", "full-ai-verification.json"))
	if err != nil {
		t.Fatalf("read full-ai verification: %v", err)
	}
	var verification struct {
		SchemaVersion string `json:"schema_version"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(verificationContents, &verification); err != nil {
		t.Fatalf("unmarshal full-ai verification: %v", err)
	}
	if verification.SchemaVersion != fullai.VerificationSchemaVersion {
		t.Fatalf("expected full-ai verification schema version %q, got %#v", fullai.VerificationSchemaVersion, verification)
	}
}

func TestAnalyzeExecutesFullAIFunctionsWithProvider(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		var payload struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode provider request: %v", err)
		}

		content := `{"summary":"Function summary","key_findings":[{"claim":"Evidence-backed claim","confidence":"high"}],"recommendations":["Keep exploring from the entry point"],"uncertainties":[]}`
		if len(payload.Messages) > 1 && strings.Contains(payload.Messages[1].Content, "Summarize this repository context") {
			content = `{"project_summary":"AI summary","architecture_narrative":"AI architecture","hotspot_explanations":[],"reading_path_explanations":[]}`
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": content}},
			},
		})
	}))
	defer providerServer.Close()

	service := New(config.Settings{
		DefaultOutputRoot: t.TempDir(),
		AppVersion:        "test",
		ConfigSource:      "test",
		Provider: config.ProviderSettings{
			Name:    "openai-compatible",
			Model:   "mistral-small-latest",
			APIKey:  "test-key",
			BaseURL: providerServer.URL,
		},
	})

	result, err := service.Analyze(t.Context(), AnalyzeRequest{
		RepoPath: filepath.Join("..", "..", "testdata", "sample-repo"),
		FullAI:   fullai.Options{Mode: fullai.ModeFull, ReadBudget: 3},
	})
	if err != nil {
		t.Fatalf("analyze sample repo with full-ai provider execution: %v", err)
	}

	if result.FullAI.Status != "executed" {
		t.Fatalf("expected executed full-ai summary, got %#v", result.FullAI)
	}
	if result.FullAI.ExecutedFunctions == 0 || result.FullAI.VerifiedFunctions == 0 {
		t.Fatalf("expected executed and verified function counts, got %#v", result.FullAI)
	}
	if result.AI.Status != "succeeded" || !result.AI.Used {
		t.Fatalf("expected legacy AI summary to be projected from full-ai output, got %#v", result.AI)
	}
	if got := int(requestCount.Load()); got != result.FullAI.PreparedFunctions {
		t.Fatalf("expected provider call per full-ai function only, got %d calls for %d functions", got, result.FullAI.PreparedFunctions)
	}
}

func TestFullAIExecutionStatusTreatsUnverifiedOutputsAsExecuted(t *testing.T) {
	t.Parallel()

	status := fullAIExecutionStatus(fullai.Execution{
		ExecutedCount: 2,
		VerifiedCount: 1,
		FailedCount:   0,
	})
	if status != "executed" {
		t.Fatalf("expected executed status when all jobs ran without failures, got %q", status)
	}
}
