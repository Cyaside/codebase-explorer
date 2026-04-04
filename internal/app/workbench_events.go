package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s Service) handleWorkbenchAnalyzeRunEvents(w http.ResponseWriter, r *http.Request, runtime *workbenchRuntime, runID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	updates, unsubscribe, ok := runtime.subscribe(runID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	defer unsubscribe()

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case run, ok := <-updates:
			if !ok {
				return
			}
			if err := writeWorkbenchEvent(w, "run", map[string]workbenchAnalyzeRun{"run": run}); err != nil {
				return
			}
			flusher.Flush()
			if isWorkbenchRunTerminal(run.Status) {
				return
			}
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func writeWorkbenchEvent(w io.Writer, eventName string, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, body)
	return err
}
