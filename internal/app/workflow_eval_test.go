package app

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Cyaside/codebase-explorer/internal/fullai"
	"github.com/Cyaside/codebase-explorer/internal/provider"
)

type fixtureStrategyMetrics struct {
	name             string
	requests         int
	verifiedSections int
	badEvidence      int
	promptBytes      int
	promptTokens     int
	outputTokens     int
	latencies        []time.Duration
	runLatencies     []time.Duration
}

func TestWorkflowStrategyFixtureComparison(t *testing.T) {
	const evidencePath = "README.md"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Messages) != 2 {
			http.Error(w, "bad fixture request", http.StatusBadRequest)
			return
		}
		user := request.Messages[1].Content
		var output any
		if strings.Contains(user, `"functions"`) {
			var batch struct {
				Functions []json.RawMessage `json:"functions"`
			}
			_ = json.Unmarshal([]byte(user), &batch)
			time.Sleep(10*time.Millisecond + time.Duration(len(batch.Functions))*12*time.Millisecond)
			output = fixtureBatchOutput(user, evidencePath)
		} else {
			time.Sleep(22 * time.Millisecond)
			output = fixtureFunctionOutput(evidencePath)
		}
		content, _ := json.Marshal(output)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]string{"content": string(content)}}},
			"usage":   map[string]int{"prompt_tokens": 100, "completion_tokens": 50},
		})
	}))
	defer server.Close()
	config := provider.Config{Name: "compatible", Model: "fixture", APIKey: "fixture-key", BaseURL: server.URL}
	client, err := provider.NewRegistry().ClientFor(config)
	if err != nil {
		t.Fatal(err)
	}
	jobs := make([]fullai.FunctionJob, 0, 7)
	for _, name := range []string{"summary", "architecture", "hotspots-and-dependencies", "flowchart", "issues", "recommendations", "dashboard"} {
		jobs = append(jobs, fullai.FunctionJob{Name: name, InstructionPath: ".agents/ai/functions/" + name + ".md", EvidencePaths: []string{evidencePath}})
	}
	evidence := fullai.Evidence{Items: []fullai.EvidenceItem{{DisplayPath: evidencePath, Snippet: "module entry point import request handler", ReadStatus: "read"}}}
	var results []fixtureStrategyMetrics
	for _, name := range []string{"one", "two", "seven"} {
		metrics := fixtureStrategyMetrics{name: name}
		for range 5 {
			cycleStarted := time.Now()
			var groups [][]fullai.FunctionJob
			switch name {
			case "one":
				groups = [][]fullai.FunctionJob{jobs[:6]}
			case "two":
				groups = [][]fullai.FunctionJob{{jobs[0], jobs[1], jobs[2], jobs[5]}, {jobs[3], jobs[4]}}
			default:
				for _, job := range jobs {
					groups = append(groups, []fullai.FunctionJob{job})
				}
			}
			if name == "two" {
				var wg sync.WaitGroup
				var mu sync.Mutex
				for _, group := range groups {
					wg.Add(1)
					go func(group []fullai.FunctionJob) {
						defer wg.Done()
						part := evaluateFixtureGroup(t, client, config, group, evidence, false)
						mu.Lock()
						metrics.merge(part)
						mu.Unlock()
					}(group)
				}
				wg.Wait()
				metrics.runLatencies = append(metrics.runLatencies, time.Since(cycleStarted))
				continue
			}
			for _, group := range groups {
				metrics.merge(evaluateFixtureGroup(t, client, config, group, evidence, name == "seven"))
			}
			metrics.runLatencies = append(metrics.runLatencies, time.Since(cycleStarted))
		}
		results = append(results, metrics)
		t.Logf("%s: requests=%d checked=%d bad_evidence=%d prompt_bytes=%d tokens_in/out=%d/%d run_p50=%s run_p95=%s request_p50=%s request_p95=%s", name,
			metrics.requests, metrics.verifiedSections, metrics.badEvidence, metrics.promptBytes, metrics.promptTokens, metrics.outputTokens,
			percentileDuration(metrics.runLatencies, .50), percentileDuration(metrics.runLatencies, .95),
			percentileDuration(metrics.latencies, .50), percentileDuration(metrics.latencies, .95))
	}
	if results[0].requests != 5 || results[1].requests != 10 || results[2].requests != 35 {
		t.Fatalf("unexpected strategy request counts: %#v", results)
	}
	for _, item := range results {
		if item.badEvidence != 0 || item.verifiedSections < 30 {
			t.Fatalf("fixture quality regressed for %s: %#v", item.name, item)
		}
	}
}

func fixtureFunctionOutput(path string) map[string]any {
	return map[string]any{
		"summary": "Fixture analysis", "key_findings": []any{map[string]any{"claim": "Fixture finding", "evidence_paths": []string{path}}},
		"recommendations": []string{"Inspect the referenced file."},
		"graph_edges":     []any{map[string]any{"from": path, "to": path, "evidence_paths": []string{path}}},
		"issue_signals":   []any{map[string]any{"title": "Fixture signal", "severity": "low", "evidence_paths": []string{path}}},
	}
}

func evaluateFixtureGroup(t *testing.T, client provider.Client, config provider.Config, jobs []fullai.FunctionJob, evidence fullai.Evidence, legacy bool) fixtureStrategyMetrics {
	t.Helper()
	metrics := fixtureStrategyMetrics{}
	if legacy {
		job := jobs[0]
		prompt, err := fullai.BuildFunctionPrompt(job, evidence)
		if err != nil {
			t.Errorf("build baseline prompt: %v", err)
			return metrics
		}
		started := time.Now()
		response, err := client.Complete(t.Context(), provider.PromptRequest{Config: config, SystemPrompt: prompt.SystemPrompt, UserPrompt: prompt.UserPrompt})
		metrics.latencies = append(metrics.latencies, time.Since(started))
		metrics.requests++
		metrics.promptBytes += len(prompt.SystemPrompt) + len(prompt.UserPrompt)
		metrics.promptTokens += response.PromptTokens
		metrics.outputTokens += response.OutputTokens
		if err != nil {
			t.Errorf("baseline request: %v", err)
			return metrics
		}
		output, err := fullai.ParseFunctionOutput(response.Content)
		if err != nil {
			t.Errorf("parse baseline output: %v", err)
			return metrics
		}
		_, report := fullai.VerifyFunctionOutputDetailed(job, output)
		if report.Verified {
			metrics.verifiedSections++
		}
		metrics.badEvidence += len(report.RejectedEvidencePaths)
		return metrics
	}
	started := time.Now()
	outcome := runAIBatch(t.Context(), client, config, jobs, evidence, workerEvidenceLimit)
	metrics.latencies = append(metrics.latencies, time.Since(started))
	metrics.requests++
	metrics.promptBytes += outcome.promptBytes
	metrics.promptTokens += outcome.promptTokens
	metrics.outputTokens += outcome.outputTokens
	if outcome.err != nil {
		t.Errorf("batch request: %v", outcome.err)
	}
	for _, result := range outcome.results {
		if result.Verified {
			metrics.verifiedSections++
		}
		metrics.badEvidence += len(result.Verification.RejectedEvidencePaths)
	}
	return metrics
}

func (metrics *fixtureStrategyMetrics) merge(other fixtureStrategyMetrics) {
	metrics.requests += other.requests
	metrics.verifiedSections += other.verifiedSections
	metrics.badEvidence += other.badEvidence
	metrics.promptBytes += other.promptBytes
	metrics.promptTokens += other.promptTokens
	metrics.outputTokens += other.outputTokens
	metrics.latencies = append(metrics.latencies, other.latencies...)
}

func percentileDuration(samples []time.Duration, quantile float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	values := slices.Clone(samples)
	slices.Sort(values)
	index := int(math.Ceil(float64(len(values))*quantile)) - 1
	return values[max(0, min(index, len(values)-1))]
}
