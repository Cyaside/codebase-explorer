import type { ReactNode, RefObject } from "react";
import { FolderSearch, LoaderCircle, OctagonX, Play, ScanSearch } from "lucide-react";

import type { AnalyzeFormState, AnalyzeRun } from "@/lib/types";

interface AnalyzeFormProps {
  activeProfileLabel: string;
  form: AnalyzeFormState;
  onChange: (next: AnalyzeFormState) => void;
  onSubmit: () => void;
  onCancel: () => void;
  busy: boolean;
  busyDetail: string;
  repoInputRef: RefObject<HTMLInputElement | null>;
  run: AnalyzeRun | null;
}

export function AnalyzeForm({ activeProfileLabel, form, onChange, onSubmit, onCancel, busy, busyDetail, repoInputRef, run }: AnalyzeFormProps) {
  const recentProgress = run?.progress.slice(-6).reverse() || [];
  const supportFileCount = countLines(form.supportFiles);
  const ignorePatternCount = countLines(form.ignorePatterns);
  const repoReady = Boolean(form.repoPath.trim());

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
        <div className="grid gap-3 md:grid-cols-3">
          <SurfaceFact
            icon={<FolderSearch className="size-4 text-primary" />}
            label="Repository"
            note={repoReady ? "Path ready for local scan" : "Waiting for a local checkout path"}
            value={repoReady ? "Ready" : "Missing"}
          />
          <SurfaceFact
            icon={<ScanSearch className="size-4 text-primary" />}
            label="Support files"
            note="Issue notes or changelog inputs"
            value={String(supportFileCount)}
          />
          <SurfaceFact
            icon={<Play className="size-4 text-primary" />}
            label="Connection"
            note="Active profile for this run"
            value={activeProfileLabel}
          />
        </div>

        <label className="field">
          <span className="field-label">Repository path</span>
          <span className="field-description">Use a local checkout path. Press <kbd className="kbd-chip">A</kbd> or <kbd className="kbd-chip">/</kbd> to focus this field instantly.</span>
          <input
            className="field-input"
            onChange={(event) => onChange({ ...form, repoPath: event.target.value })}
            placeholder="C:\\projects\\my-repo"
            ref={repoInputRef}
            type="text"
            value={form.repoPath}
          />
        </label>

        <div className="grid gap-4 xl:grid-cols-2">
          <label className="field">
            <span className="field-label">Support files</span>
            <span className="field-description">One path per line for issue exports, changelogs, or handwritten notes that should shape change-awareness.</span>
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
            <span className="field-description">Optional extra ignore rules for local noise. Keep these narrow so the analyzer still sees the real project shape.</span>
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
          <div className="space-y-2 text-sm text-muted-foreground">
            <p>GitHub URLs still intentionally stay out of scope here. Keep analysis local for predictable, lightweight runs.</p>
            <div className="flex flex-wrap gap-2 text-xs">
              <ShortcutHint keys="Ctrl+Enter" label="Analyze" />
              <ShortcutHint keys="Ctrl+K" label="Palette" />
              <ShortcutHint keys="Esc" label="Cancel run" />
              <ShortcutHint keys="[" label="Prev tab" />
              <ShortcutHint keys="]" label="Next tab" />
              <ShortcutHint keys="J / K" label="Switch bundle" />
              <ShortcutHint keys="R" label="Refresh" />
            </div>
          </div>
          <div className="flex flex-wrap gap-3">
            <button className="action-button" disabled={busy} onClick={onSubmit} type="button">
              {busy ? <LoaderCircle className="size-4 animate-spin" /> : <Play className="size-4" />}
              {busy ? busyDetail || "Analyzing..." : "Analyze Project"}
            </button>
            {busy ? (
              <button className="secondary-button danger" onClick={onCancel} type="button">
                <OctagonX className="size-4" />
                Cancel Run
              </button>
            ) : null}
          </div>
        </div>

        {run ? (
          <div className="rounded-2xl border border-border bg-background/70 p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Analyze progress</p>
                <p className="mt-1 text-sm text-foreground">
                  Run <span className="font-mono text-xs text-muted-foreground">{run.id}</span>
                </p>
              </div>
              <span
                className={
                  run.status === "failed" || run.status === "canceled"
                    ? "status-pill border-danger/35 bg-danger/10 text-danger"
                    : busy
                      ? "status-pill status-pill-active"
                      : "status-pill status-pill-neutral"
                }
              >
                {run.status}
              </span>
            </div>

            <div className="space-y-2">
              {recentProgress.length ? (
                recentProgress.map((event, index) => (
                  <article className="rounded-2xl border border-border bg-card/70 px-3 py-3" key={`${event.stage}-${event.status}-${index}`}>
                    <div className="flex flex-wrap items-center gap-2 text-xs uppercase tracking-[0.14em] text-muted-foreground">
                      <span>{event.stage || "stage"}</span>
                      <span>/</span>
                      <span>{event.status || "status"}</span>
                    </div>
                    <p className="mt-2 text-sm text-foreground">{event.detail || "No extra detail."}</p>
                  </article>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">Waiting for progress events from the backend.</p>
              )}
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}

function SurfaceFact({
  icon,
  label,
  note,
  value,
}: {
  icon: ReactNode;
  label: string;
  note: string;
  value: string;
}) {
  return (
    <article className="surface-fact">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{label}</span>
        {icon}
      </div>
      <strong className="mt-3 block text-base font-semibold text-foreground">{value}</strong>
      <p className="mt-1 text-xs text-muted-foreground">{note}</p>
    </article>
  );
}

function ShortcutHint({ keys, label }: { keys: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-border bg-background/60 px-2.5 py-1">
      <kbd className="kbd-chip">{keys}</kbd>
      <span>{label}</span>
    </span>
  );
}

function countLines(value: string) {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean).length;
}
