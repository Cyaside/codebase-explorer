package fullai

import (
	"strings"
	"testing"
	"testing/fstest"
)

func fixturePack() fstest.MapFS {
	return fstest.MapFS{
		".agents/README.md":               {Data: []byte("pack overview")},
		".agents/agents.md":               {Data: []byte("worker contract")},
		".agents/ai/README.md":            {Data: []byte("orchestrator")},
		".agents/ai/functions/summary.md": {Data: []byte("summary spec")},
	}
}

func TestInstructionsLoadPackInRequiredOrderAndChangeHash(t *testing.T) {
	t.Parallel()
	files := fixturePack()
	job := FunctionJob{Name: "summary", InstructionPath: ".agents/ai/functions/summary.md"}
	first, err := loadInstructionsFrom(files, job)
	if err != nil {
		t.Fatalf("load instructions: %v", err)
	}
	for _, marker := range []string{"pack overview", "worker contract", "orchestrator", "summary spec"} {
		if !strings.Contains(first.Text, marker) {
			t.Fatalf("expected instruction %q in prompt, got %q", marker, first.Text)
		}
	}
	if !(strings.Index(first.Text, "pack overview") < strings.Index(first.Text, "worker contract") &&
		strings.Index(first.Text, "worker contract") < strings.Index(first.Text, "orchestrator") &&
		strings.Index(first.Text, "orchestrator") < strings.Index(first.Text, "summary spec")) {
		t.Fatalf("instruction order is wrong: %q", first.Text)
	}
	files[".agents/ai/functions/summary.md"] = &fstest.MapFile{Data: []byte("revised summary spec")}
	second, err := loadInstructionsFrom(files, job)
	if err != nil {
		t.Fatalf("reload instructions: %v", err)
	}
	if first.Hash == second.Hash || !strings.Contains(second.Text, "revised summary spec") {
		t.Fatalf("changing the guide must change the prompt and hash")
	}
}

func TestInstructionsRejectMissingOrMismatchedGuide(t *testing.T) {
	t.Parallel()
	files := fixturePack()
	delete(files, ".agents/agents.md")
	if _, err := loadInstructionsFrom(files, FunctionJob{Name: "summary"}); err == nil {
		t.Fatal("missing global guide must fail")
	}
	if _, err := loadInstructionsFrom(fixturePack(), FunctionJob{Name: "summary", InstructionPath: ".agents/ai/functions/issues.md"}); err == nil {
		t.Fatal("mismatched function guide must fail")
	}
}
