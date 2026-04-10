package fullai

import "testing"

func TestBuildFunctionPromptIncludesEvidence(t *testing.T) {
	t.Parallel()

	prompt, err := BuildFunctionPrompt(FunctionJob{
		Name:          "summary",
		Objective:     "summarize",
		EvidencePaths: []string{"cmd/main.go"},
	}, Evidence{
		Items: []EvidenceItem{
			{DisplayPath: "cmd/main.go", Snippet: "package main"},
		},
	})
	if err != nil {
		t.Fatalf("build function prompt: %v", err)
	}
	if prompt.FunctionName != "summary" || prompt.SystemPrompt == "" || prompt.UserPrompt == "" {
		t.Fatalf("expected populated prompt, got %#v", prompt)
	}
}

func TestParseFunctionOutputStripsFence(t *testing.T) {
	t.Parallel()

	output, err := ParseFunctionOutput("```json\n{\"summary\":\"ok\",\"key_findings\":[{\"claim\":\"claim\",\"evidence_paths\":[\"cmd/main.go\"]}]}\n```")
	if err != nil {
		t.Fatalf("parse function output: %v", err)
	}
	if output.Summary != "ok" || len(output.KeyFindings) != 1 {
		t.Fatalf("expected parsed function output, got %#v", output)
	}
}

func TestVerifyFunctionOutputFiltersUnknownEvidencePaths(t *testing.T) {
	t.Parallel()

	output, verified := VerifyFunctionOutput(FunctionJob{
		EvidencePaths: []string{"cmd/main.go"},
	}, FunctionOutput{
		KeyFindings: []Finding{
			{Claim: "claim", EvidencePaths: []string{"cmd/main.go", "missing.go"}},
		},
	})
	if verified {
		t.Fatalf("expected verifier to flag unknown evidence path")
	}
	if len(output.KeyFindings[0].EvidencePaths) != 1 {
		t.Fatalf("expected unknown path to be filtered, got %#v", output)
	}
}
