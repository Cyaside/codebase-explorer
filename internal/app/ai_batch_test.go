package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/analyzer"
	"github.com/Cyaside/codebase-explorer/internal/config"
	"github.com/Cyaside/codebase-explorer/internal/fullai"
)

func TestAdaptiveAIUsesTwoParallelCallsAndLoadsAgentPack(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	active, peak, calls := 0, 0, 0
	secondCall := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Messages) < 2 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if !strings.Contains(payload.Messages[0].Content, "# Instruction: .agents/agents.md") ||
			!strings.Contains(payload.Messages[0].Content, "# Instruction: .agents/ai/functions/summary.md") &&
				!strings.Contains(payload.Messages[0].Content, "# Instruction: .agents/ai/functions/flowchart.md") {
			http.Error(w, "agent pack missing", http.StatusBadRequest)
			return
		}
		mu.Lock()
		active++
		calls++
		if active > peak {
			peak = active
		}
		if active == 2 {
			close(secondCall)
		}
		mu.Unlock()
		select {
		case <-secondCall:
		case <-time.After(time.Second):
		}
		mu.Lock()
		active--
		mu.Unlock()
		content, _ := json.Marshal(fixtureBatchOutput(payload.Messages[1].Content, "README.md"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
	}))
	defer server.Close()
	service := New(config.Settings{AppVersion: "test", DefaultOutputRoot: t.TempDir(), Provider: config.ProviderSettings{
		Name: "compatible", Model: "fixture", APIKey: "test-key", BaseURL: server.URL,
	}})
	jobs := []fullai.FunctionJob{}
	for _, name := range []string{"summary", "architecture", "hotspots-and-dependencies", "flowchart", "issues", "recommendations", "dashboard"} {
		jobs = append(jobs, fullai.FunctionJob{Name: name, InstructionPath: ".agents/ai/functions/" + name + ".md", EvidencePaths: []string{"README.md"}})
	}
	evidence := fullai.Evidence{Items: []fullai.EvidenceItem{{DisplayPath: "README.md", Snippet: strings.Repeat("code\n", 5000), ReadStatus: "read"}}}
	execution, _, err := service.executeAdaptiveAI(t.Context(), AnalyzeRequest{}, analyzer.Result{}, fullai.Functions{Jobs: jobs}, evidence, fullai.Summary{})
	if err != nil {
		t.Fatalf("execute adaptive AI: %v", err)
	}
	if execution.Status != "executed" || execution.PackHash == "" {
		t.Fatalf("expected verified execution with pack hash, got %#v", execution)
	}
	mu.Lock()
	defer mu.Unlock()
	if calls != 2 || peak != 2 {
		t.Fatalf("expected two concurrent calls, got calls=%d peak=%d", calls, peak)
	}
}

func TestAdaptiveGroupingUsesEvidenceAndProjectedOutputSize(t *testing.T) {
	t.Parallel()
	jobs := []fullai.FunctionJob{{Name: "summary"}, {Name: "architecture"}, {Name: "hotspots-and-dependencies"}, {Name: "flowchart"}, {Name: "issues"}, {Name: "recommendations"}}
	compact := fullai.Evidence{Items: []fullai.EvidenceItem{{Snippet: strings.Repeat("x", 17900)}}}
	large := fullai.Evidence{Items: []fullai.EvidenceItem{{Snippet: strings.Repeat("x", 18100)}}}
	if len(groupAIJobs(jobs, compact)) != 1 || len(groupAIJobs(jobs, large)) != 2 {
		t.Fatal("adaptive grouping did not account for combined evidence and projected output size")
	}
}

func TestRepairOnlyRequestsTheMissingSection(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	var repairNames []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		var requested struct {
			Functions []struct {
				Name string `json:"name"`
			} `json:"functions"`
		}
		_ = json.Unmarshal([]byte(payload.Messages[1].Content), &requested)
		fixture := fixtureBatchOutput(payload.Messages[1].Content, "README.md")
		if calls.Add(1) == 1 {
			delete(fixture["functions"].(map[string]any), "issues")
		} else {
			for _, job := range requested.Functions {
				repairNames = append(repairNames, job.Name)
			}
		}
		content, _ := json.Marshal(fixture)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}}})
	}))
	defer server.Close()
	service := New(config.Settings{AppVersion: "test", DefaultOutputRoot: t.TempDir(), Provider: config.ProviderSettings{
		Name: "compatible", Model: "fixture", APIKey: "test-key", BaseURL: server.URL,
	}})
	jobs := make([]fullai.FunctionJob, 0, 7)
	for _, name := range []string{"summary", "architecture", "hotspots-and-dependencies", "flowchart", "issues", "recommendations", "dashboard"} {
		jobs = append(jobs, fullai.FunctionJob{Name: name, InstructionPath: ".agents/ai/functions/" + name + ".md", EvidencePaths: []string{"README.md"}})
	}
	evidence := fullai.Evidence{Items: []fullai.EvidenceItem{{DisplayPath: "README.md", Snippet: "project evidence", ReadStatus: "read"}}}
	execution, _, err := service.executeAdaptiveAI(t.Context(), AnalyzeRequest{}, analyzer.Result{}, fullai.Functions{Jobs: jobs}, evidence, fullai.Summary{})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Status != "executed" || execution.RepairCalls != 1 || calls.Load() != 2 || len(repairNames) != 1 || repairNames[0] != "issues" {
		t.Fatalf("repair was not scoped to missing issues: status=%s repair=%d calls=%d names=%v", execution.Status, execution.RepairCalls, calls.Load(), repairNames)
	}
}
