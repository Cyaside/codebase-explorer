import { LoaderCircle, Play } from "lucide-react";

import type { AnalyzeFormState } from "@/lib/types";

interface AnalyzeFormProps {
  form: AnalyzeFormState;
  onChange: (next: AnalyzeFormState) => void;
  onSubmit: () => void;
  busy: boolean;
  busyDetail: string;
}

export function AnalyzeForm({ form, onChange, onSubmit, busy, busyDetail }: AnalyzeFormProps) {
  return (
    <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-sm">
      <div className="mb-5 flex items-start justify-between gap-3">
        <div>
          <p className="text-[10px] uppercase tracking-[0.28em] text-primary">Analyze</p>
          <h2 className="text-xl font-bold">Open a local project</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Point the workbench at a local checkout, optionally attach issue notes, and produce a reusable orientation bundle.
          </p>
        </div>
        <span className={busy ? "status-pill status-pill-active" : "status-pill status-pill-neutral"}>
          {busy ? "Running" : "Idle"}
        </span>
      </div>

      <div className="space-y-4">
        <label className="field">
          <span className="field-label">Repository path</span>
          <input
            className="field-input"
            onChange={(event) => onChange({ ...form, repoPath: event.target.value })}
            placeholder="C:\\projects\\my-repo"
            type="text"
            value={form.repoPath}
          />
        </label>

        <div className="grid gap-4 xl:grid-cols-2">
          <label className="field">
            <span className="field-label">Support files</span>
            <textarea
              className="field-textarea"
              onChange={(event) => onChange({ ...form, supportFiles: event.target.value })}
              placeholder="One path per line for issues, changelog, or notes"
              rows={5}
              value={form.supportFiles}
            />
          </label>

          <label className="field">
            <span className="field-label">Ignore patterns</span>
            <textarea
              className="field-textarea"
              onChange={(event) => onChange({ ...form, ignorePatterns: event.target.value })}
              placeholder="Optional extra ignore patterns, one per line"
              rows={5}
              value={form.ignorePatterns}
            />
          </label>
        </div>

        <div className="flex flex-col gap-3 border-t border-border pt-4 md:flex-row md:items-center md:justify-between">
          <div className="text-sm text-muted-foreground">
            GitHub URLs still intentionally stay out of scope here. Keep analysis local for predictable, lightweight runs.
          </div>
          <button className="action-button" disabled={busy} onClick={onSubmit} type="button">
            {busy ? <LoaderCircle className="size-4 animate-spin" /> : <Play className="size-4" />}
            {busy ? busyDetail || "Analyzing..." : "Analyze Project"}
          </button>
        </div>
      </div>
    </section>
  );
}
