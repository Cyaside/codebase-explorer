package app

import (
	"testing"
)

func TestWorkbenchRejectsThirdActiveRun(t *testing.T) {
	t.Parallel()
	runtime := newWorkbenchRuntime(Service{}, "")
	runtime.runs["one"] = &workbenchRunState{run: workbenchAnalyzeRun{Status: "running"}}
	runtime.runs["two"] = &workbenchRunState{run: workbenchAnalyzeRun{Status: "canceling"}}
	if _, err := runtime.startAnalyze(workbenchAnalyzePayload{}); err == nil {
		t.Fatal("expected active run limit")
	}
}

func TestWorkbenchTrimKeepsActiveRuns(t *testing.T) {
	t.Parallel()
	runtime := newWorkbenchRuntime(Service{}, "")
	for i := 0; i < 13; i++ {
		id := string(rune('a' + i))
		status := "succeeded"
		if i == 0 {
			status = "running"
		}
		runtime.runs[id] = &workbenchRunState{run: workbenchAnalyzeRun{Status: status}}
		runtime.runList = append(runtime.runList, id)
	}
	runtime.trimRunsLocked()
	if _, exists := runtime.runs["a"]; !exists {
		t.Fatal("active run was trimmed")
	}
	if len(runtime.runList) != 12 {
		t.Fatalf("expected 12 retained runs, got %d", len(runtime.runList))
	}
}
