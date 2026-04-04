import type { AnalyzeRun } from "@/lib/types";
import { normalizeAnalyzeRun } from "@/lib/normalize";

interface AnalyzeRunStreamHandlers {
  onRun: (run: AnalyzeRun) => void;
  onFallback: () => void;
}

function isTerminalRunStatus(status: string) {
  return status === "succeeded" || status === "failed" || status === "canceled";
}

export function openAnalyzeRunStream(runID: string, handlers: AnalyzeRunStreamHandlers) {
  if (typeof window === "undefined" || typeof window.EventSource === "undefined") {
    handlers.onFallback();
    return () => {};
  }

  let closed = false;
  const events = new window.EventSource(`/api/analyze-runs/${encodeURIComponent(runID)}/events`);

  const close = () => {
    if (closed) {
      return;
    }
    closed = true;
    events.close();
  };

  events.addEventListener("run", (event) => {
    try {
      const payload = JSON.parse((event as MessageEvent<string>).data) as { run: AnalyzeRun };
      const run = normalizeAnalyzeRun(payload.run);
      handlers.onRun(run);
      if (isTerminalRunStatus(run.status)) {
        close();
      }
    } catch {
      close();
      handlers.onFallback();
    }
  });

  events.onerror = () => {
    if (closed) {
      return;
    }
    close();
    handlers.onFallback();
  };

  return close;
}
