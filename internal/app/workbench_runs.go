package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const maxWorkbenchProgressEvents = 64

type workbenchAnalyzeRun struct {
	ID        string                    `json:"id"`
	Status    string                    `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
	Progress  []AnalyzeProgressEvent    `json:"progress"`
	Error     string                    `json:"error,omitempty"`
	Response  *workbenchAnalyzeResponse `json:"response,omitempty"`
}

type workbenchRunState struct {
	run         workbenchAnalyzeRun
	cancel      context.CancelFunc
	subscribers map[int]chan workbenchAnalyzeRun
}

type workbenchRuntime struct {
	service    Service
	outputRoot string

	mu      sync.Mutex
	seq     int
	subSeq  int
	runs    map[string]*workbenchRunState
	runList []string
}

func newWorkbenchRuntime(service Service, outputRoot string) *workbenchRuntime {
	return &workbenchRuntime{
		service:    service,
		outputRoot: outputRoot,
		runs:       map[string]*workbenchRunState{},
	}
}

func (runtime *workbenchRuntime) startAnalyze(payload workbenchAnalyzePayload) workbenchAnalyzeRun {
	runtime.mu.Lock()
	runtime.seq++
	runID := fmt.Sprintf("run-%d-%03d", time.Now().UTC().Unix(), runtime.seq)
	now := time.Now().UTC()
	state := &workbenchRunState{
		run: workbenchAnalyzeRun{
			ID:        runID,
			Status:    "queued",
			CreatedAt: now,
			UpdatedAt: now,
			Progress:  []AnalyzeProgressEvent{{Stage: "analyze", Status: "queued", Detail: "analysis queued in local workbench"}},
		},
	}
	runtime.runs[runID] = state
	runtime.runList = append(runtime.runList, runID)
	closedSubscribers := runtime.trimRunsLocked()
	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	runtime.mu.Unlock()
	closeWorkbenchRunSubscribers(closedSubscribers)

	go runtime.executeAnalyze(runID, payload)

	return snapshot
}

func (runtime *workbenchRuntime) executeAnalyze(runID string, payload workbenchAnalyzePayload) {
	ctx, cancel := context.WithCancel(context.Background())
	runtime.setRunCancel(runID, cancel)
	runtime.appendProgress(runID, AnalyzeProgressEvent{
		Stage:  "analyze",
		Status: "running",
		Detail: "analysis started in background",
	})

	result, err := runtime.service.Analyze(ctx, AnalyzeRequest{
		RepoPath:             payload.RepoPath,
		DeterministicOnly:    payload.DeterministicOnly,
		ExtraIgnorePatterns:  payload.ExtraIgnorePatterns,
		OptionalSupportFiles: payload.SupportFiles,
		ProviderOverride:     payload.Provider,
		Progress: func(event AnalyzeProgressEvent) {
			runtime.appendProgress(runID, event)
		},
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			runtime.finishCanceled(runID, "analysis canceled by user")
			return
		}
		runtime.finishFailed(runID, err.Error())
		return
	}

	response, err := runtime.service.buildWorkbenchAnalyzeResponse(result)
	if err != nil {
		runtime.finishFailed(runID, err.Error())
		return
	}

	runtime.finishSucceeded(runID, response)
}

func (runtime *workbenchRuntime) snapshot(runID string) (workbenchAnalyzeRun, bool) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	state, ok := runtime.runs[runID]
	if !ok {
		return workbenchAnalyzeRun{}, false
	}

	return cloneWorkbenchAnalyzeRun(state.run), true
}

func (runtime *workbenchRuntime) cancel(runID string) (workbenchAnalyzeRun, bool) {
	runtime.mu.Lock()
	state, ok := runtime.runs[runID]
	if !ok {
		runtime.mu.Unlock()
		return workbenchAnalyzeRun{}, false
	}

	if state.cancel != nil && (state.run.Status == "queued" || state.run.Status == "running") {
		state.run.Status = "canceling"
		state.run.UpdatedAt = time.Now().UTC()
		state.run.Progress = appendProgressEvent(state.run.Progress, AnalyzeProgressEvent{
			Stage:  "analyze",
			Status: "canceling",
			Detail: "cancel requested from workbench",
		})
		cancel := state.cancel
		snapshot := cloneWorkbenchAnalyzeRun(state.run)
		subscribers := cloneWorkbenchRunSubscribers(state)
		runtime.mu.Unlock()
		broadcastWorkbenchRun(snapshot, subscribers)
		cancel()
		return snapshot, true
	}

	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	runtime.mu.Unlock()
	return snapshot, true
}

func (runtime *workbenchRuntime) setRunCancel(runID string, cancel context.CancelFunc) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	state, ok := runtime.runs[runID]
	if !ok {
		return
	}

	state.cancel = cancel
	if state.run.Status == "queued" {
		state.run.Status = "running"
		state.run.UpdatedAt = time.Now().UTC()
	}
}

func (runtime *workbenchRuntime) appendProgress(runID string, event AnalyzeProgressEvent) {
	runtime.mu.Lock()
	state, ok := runtime.runs[runID]
	if !ok {
		runtime.mu.Unlock()
		return
	}

	if state.run.Status == "queued" {
		state.run.Status = "running"
	}
	state.run.UpdatedAt = time.Now().UTC()
	state.run.Progress = appendProgressEvent(state.run.Progress, event)
	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	subscribers := cloneWorkbenchRunSubscribers(state)
	runtime.mu.Unlock()

	broadcastWorkbenchRun(snapshot, subscribers)
}

func (runtime *workbenchRuntime) finishSucceeded(runID string, response workbenchAnalyzeResponse) {
	runtime.mu.Lock()
	state, ok := runtime.runs[runID]
	if !ok {
		runtime.mu.Unlock()
		return
	}

	state.cancel = nil
	state.run.Status = "succeeded"
	state.run.UpdatedAt = time.Now().UTC()
	state.run.Progress = appendProgressEvent(state.run.Progress, AnalyzeProgressEvent{
		Stage:  "analyze",
		Status: "succeeded",
		Detail: "background analysis completed",
	})
	state.run.Response = &response
	state.run.Error = ""
	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	subscribers := takeWorkbenchRunSubscribers(state)
	runtime.mu.Unlock()

	broadcastWorkbenchRun(snapshot, subscribers)
	closeWorkbenchRunSubscribers(subscribers)
}

func (runtime *workbenchRuntime) finishFailed(runID string, message string) {
	runtime.mu.Lock()
	state, ok := runtime.runs[runID]
	if !ok {
		runtime.mu.Unlock()
		return
	}

	state.cancel = nil
	state.run.Status = "failed"
	state.run.UpdatedAt = time.Now().UTC()
	state.run.Progress = appendProgressEvent(state.run.Progress, AnalyzeProgressEvent{
		Stage:  "analyze",
		Status: "failed",
		Detail: message,
	})
	state.run.Error = message
	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	subscribers := takeWorkbenchRunSubscribers(state)
	runtime.mu.Unlock()

	broadcastWorkbenchRun(snapshot, subscribers)
	closeWorkbenchRunSubscribers(subscribers)
}

func (runtime *workbenchRuntime) finishCanceled(runID string, message string) {
	runtime.mu.Lock()
	state, ok := runtime.runs[runID]
	if !ok {
		runtime.mu.Unlock()
		return
	}

	state.cancel = nil
	state.run.Status = "canceled"
	state.run.UpdatedAt = time.Now().UTC()
	state.run.Progress = appendProgressEvent(state.run.Progress, AnalyzeProgressEvent{
		Stage:  "analyze",
		Status: "canceled",
		Detail: message,
	})
	state.run.Error = message
	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	subscribers := takeWorkbenchRunSubscribers(state)
	runtime.mu.Unlock()

	broadcastWorkbenchRun(snapshot, subscribers)
	closeWorkbenchRunSubscribers(subscribers)
}

func (runtime *workbenchRuntime) subscribe(runID string) (<-chan workbenchAnalyzeRun, func(), bool) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	state, ok := runtime.runs[runID]
	if !ok {
		return nil, nil, false
	}

	snapshot := cloneWorkbenchAnalyzeRun(state.run)
	if isWorkbenchRunTerminal(snapshot.Status) {
		updates := make(chan workbenchAnalyzeRun, 1)
		updates <- snapshot
		close(updates)
		return updates, func() {}, true
	}

	runtime.subSeq++
	subscriberID := runtime.subSeq
	if state.subscribers == nil {
		state.subscribers = map[int]chan workbenchAnalyzeRun{}
	}

	updates := make(chan workbenchAnalyzeRun, 8)
	state.subscribers[subscriberID] = updates
	updates <- snapshot

	return updates, func() {
		runtime.unsubscribe(runID, subscriberID)
	}, true
}

func (runtime *workbenchRuntime) unsubscribe(runID string, subscriberID int) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()

	state, ok := runtime.runs[runID]
	if !ok || state.subscribers == nil {
		return
	}

	delete(state.subscribers, subscriberID)
}

func (runtime *workbenchRuntime) trimRunsLocked() []chan workbenchAnalyzeRun {
	const keepLatestRuns = 12
	if len(runtime.runList) <= keepLatestRuns {
		return nil
	}

	excess := len(runtime.runList) - keepLatestRuns
	var subscribersToClose []chan workbenchAnalyzeRun
	for _, runID := range runtime.runList[:excess] {
		if state, ok := runtime.runs[runID]; ok {
			subscribersToClose = append(subscribersToClose, takeWorkbenchRunSubscribers(state)...)
			delete(runtime.runs, runID)
		}
	}
	runtime.runList = append([]string(nil), runtime.runList[excess:]...)
	return subscribersToClose
}

func appendProgressEvent(events []AnalyzeProgressEvent, event AnalyzeProgressEvent) []AnalyzeProgressEvent {
	events = append(events, event)
	if len(events) <= maxWorkbenchProgressEvents {
		return events
	}
	return append([]AnalyzeProgressEvent(nil), events[len(events)-maxWorkbenchProgressEvents:]...)
}

func cloneWorkbenchAnalyzeRun(run workbenchAnalyzeRun) workbenchAnalyzeRun {
	cloned := run
	cloned.Progress = append([]AnalyzeProgressEvent(nil), run.Progress...)
	return cloned
}

func cloneWorkbenchRunSubscribers(state *workbenchRunState) []chan workbenchAnalyzeRun {
	if len(state.subscribers) == 0 {
		return nil
	}

	subscribers := make([]chan workbenchAnalyzeRun, 0, len(state.subscribers))
	for _, updates := range state.subscribers {
		subscribers = append(subscribers, updates)
	}
	return subscribers
}

func takeWorkbenchRunSubscribers(state *workbenchRunState) []chan workbenchAnalyzeRun {
	subscribers := cloneWorkbenchRunSubscribers(state)
	state.subscribers = nil
	return subscribers
}

func broadcastWorkbenchRun(snapshot workbenchAnalyzeRun, subscribers []chan workbenchAnalyzeRun) {
	for _, updates := range subscribers {
		pushWorkbenchRunUpdate(updates, snapshot)
	}
}

func pushWorkbenchRunUpdate(updates chan workbenchAnalyzeRun, snapshot workbenchAnalyzeRun) {
	select {
	case updates <- snapshot:
		return
	default:
	}

	select {
	case <-updates:
	default:
	}

	select {
	case updates <- snapshot:
	default:
	}
}

func closeWorkbenchRunSubscribers(subscribers []chan workbenchAnalyzeRun) {
	for _, updates := range subscribers {
		close(updates)
	}
}

func isWorkbenchRunTerminal(status string) bool {
	switch status {
	case "succeeded", "failed", "canceled":
		return true
	default:
		return false
	}
}
