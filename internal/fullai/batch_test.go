package fullai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBatchPromptSharesEvidenceBudgetAcrossFiles(t *testing.T) {
	t.Parallel()
	job := FunctionJob{Name: "summary", InstructionPath: ".agents/ai/functions/summary.md", EvidencePaths: []string{"a.go", "b.go"}}
	prompt, err := BuildBatchPrompt([]FunctionJob{job}, Evidence{Items: []EvidenceItem{
		{DisplayPath: "a.go", Snippet: strings.Repeat("a", 2000)},
		{DisplayPath: "b.go", Snippet: strings.Repeat("b", 2000)},
	}}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		EvidenceItems []EvidenceItem `json:"evidence_items"`
	}
	if err := json.Unmarshal([]byte(prompt.UserPrompt), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.EvidenceItems) != 2 || len(payload.EvidenceItems[0].Snippet) != 500 || len(payload.EvidenceItems[1].Snippet) != 500 {
		t.Fatalf("expected balanced evidence excerpts, got %#v", payload.EvidenceItems)
	}
}

func TestBatchPromptContainsCurrentAgentGuideContents(t *testing.T) {
	t.Parallel()
	files := fixturePack()
	job := FunctionJob{Name: "summary", InstructionPath: ".agents/ai/functions/summary.md", EvidencePaths: []string{"README.md"}}
	first, err := buildBatchPromptFrom(files, []FunctionJob{job}, Evidence{}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	files[".agents/ai/functions/summary.md"].Data = []byte("changed summary guidance")
	second, err := buildBatchPromptFrom(files, []FunctionJob{job}, Evidence{}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if first.SystemPrompt == second.SystemPrompt || !strings.Contains(second.SystemPrompt, "changed summary guidance") {
		t.Fatal("batch provider payload did not change with the embedded guide content")
	}
}

func TestParseBatchOutputRepairsTrailingJSONCommas(t *testing.T) {
	t.Parallel()
	content := `{"functions":{"summary":{"summary":"comma,} stays intact","key_findings":[],},"issues":{"summary":"ok"},},}`
	outputs, err := ParseBatchOutput(content)
	if err != nil || len(outputs) != 2 || outputs["summary"].Summary != "comma,} stays intact" {
		t.Fatalf("trailing commas discarded usable sections: outputs=%#v error=%v", outputs, err)
	}
}

func TestParseBatchOutputKeepsCompletedSectionsFromBrokenResponse(t *testing.T) {
	t.Parallel()
	content := `{"functions":{"summary":{"summary":"complete"},"architecture":{"summary":"also complete"},"issues":{broken}}}`
	outputs, err := ParseBatchOutput(content)
	if err != nil || len(outputs) != 2 || outputs["summary"].Summary != "complete" || outputs["architecture"].Summary != "also complete" {
		t.Fatalf("completed functions were lost: outputs=%#v error=%v", outputs, err)
	}
}
