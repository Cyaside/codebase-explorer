import { useState, type RefObject } from "react";
import { ChevronDown, LoaderCircle, OctagonX, Play, Plus, SlidersHorizontal } from "lucide-react";

import type { AnalyzeRun, ConnectionProfile, SavedWorkspace } from "@/lib/types";
import { formatRelativeTime } from "@/lib/utils";

interface AnalyzeFormProps {
  busy: boolean;
  busyDetail: string;
  onCancel: () => void;
  onCreateWorkspace: () => void;
  onSubmit: () => void;
  onWorkspaceChange: (workspace: SavedWorkspace) => void;
  profiles: ConnectionProfile[];
  repoInputRef: RefObject<HTMLInputElement | null>;
  run: AnalyzeRun | null;
  workspace: SavedWorkspace | null;
}

export function AnalyzeForm({
  busy,
  busyDetail,
  onCancel,
  onCreateWorkspace,
  onSubmit,
  onWorkspaceChange,
  profiles,
  repoInputRef,
  run,
  workspace,
}: AnalyzeFormProps) {
  const [advancedOpen, setAdvancedOpen] = useState(false);

  if (!workspace) {
    return (
      <section className="panel-block flex min-h-[16rem] items-center justify-between gap-6">
        <div className="max-w-xl">
          <p className="panel-kicker">Launcher</p>
          <h2 className="mt-3 text-2xl font-semibold tracking-tight text-zinc-50">Open one project and keep the whole shell scoped to it.</h2>
          <p className="mt-4 text-sm leading-7 text-zinc-400">
            This workbench is no longer a global bundle browser. Create a workspace, point it at one local repository, then keep analysis,
            graph inspection, issues, and recommendations anchored to that project.
          </p>
        </div>
        <button className="primary-control" onClick={onCreateWorkspace} type="button">
          <Plus className="size-4" />
          New Project
        </button>
      </section>
    );
  }

  const recentProgress = run?.progress.slice(-4).reverse() || [];

  return (
    <section className="panel-block space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="panel-kicker">Project setup</p>
          <h2 className="mt-3 text-2xl font-semibold tracking-tight text-zinc-50">{workspace.label}</h2>
          <p className="mt-2 text-sm text-zinc-500">One active project per app instance. Saved locally so you can reopen it later.</p>
        </div>
        <div className="flex items-center gap-2">
          <span className={busy ? "run-pill run-pill-busy" : "run-pill"}>{busy ? "Running" : "Ready"}</span>
          <button className="secondary-control" onClick={() => setAdvancedOpen((current) => !current)} type="button">
            <SlidersHorizontal className="size-4" />
            Advanced
          </button>
        </div>
      </div>

      <div className="grid gap-3 xl:grid-cols-[minmax(0,16rem)_minmax(0,1fr)_minmax(0,13rem)_auto]">
        <label className="compact-field">
          <span className="compact-label">Workspace</span>
          <input
            className="compact-input"
            onChange={(event) => onWorkspaceChange({ ...workspace, label: event.target.value, updatedAt: new Date().toISOString() })}
            type="text"
            value={workspace.label}
          />
        </label>

        <label className="compact-field">
          <span className="compact-label">Repository path</span>
          <input
            className="compact-input"
            onChange={(event) => onWorkspaceChange({ ...workspace, repoPath: event.target.value, updatedAt: new Date().toISOString() })}
            placeholder="C:\\projects\\your-repo"
            ref={repoInputRef}
            type="text"
            value={workspace.repoPath}
          />
        </label>

        <label className="compact-field">
          <span className="compact-label">Connection</span>
          <select
            className="compact-input"
            onChange={(event) => onWorkspaceChange({ ...workspace, selectedProfile: event.target.value, updatedAt: new Date().toISOString() })}
            value={workspace.selectedProfile}
          >
            {profiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {profile.label}
              </option>
            ))}
          </select>
        </label>

        <div className="flex items-end gap-2">
          <button className="primary-control" disabled={busy} onClick={onSubmit} type="button">
            {busy ? <LoaderCircle className="size-4 animate-spin" /> : <Play className="size-4" />}
            {busy ? "Analyzing" : "Analyze"}
          </button>
          {busy ? (
            <button className="secondary-control danger" onClick={onCancel} type="button">
              <OctagonX className="size-4" />
              Cancel
            </button>
          ) : null}
        </div>
      </div>

      {advancedOpen ? (
        <div className="grid gap-3 border-t border-zinc-900 pt-4 xl:grid-cols-2">
          <label className="compact-field">
            <span className="compact-label">Support files</span>
            <textarea
              className="compact-area"
              onChange={(event) =>
                onWorkspaceChange({
                  ...workspace,
                  supportFiles: splitToLines(event.target.value),
                  updatedAt: new Date().toISOString(),
                })
              }
              placeholder="One local path per line"
              rows={5}
              value={workspace.supportFiles.join("\n")}
            />
          </label>
          <label className="compact-field">
            <span className="compact-label">Ignore patterns</span>
            <textarea
              className="compact-area"
              onChange={(event) =>
                onWorkspaceChange({
                  ...workspace,
                  ignorePatterns: splitToLines(event.target.value),
                  updatedAt: new Date().toISOString(),
                })
              }
              placeholder="One extra ignore pattern per line"
              rows={5}
              value={workspace.ignorePatterns.join("\n")}
            />
          </label>
        </div>
      ) : null}

      <div className="flex flex-wrap items-center justify-between gap-4 border-t border-zinc-900 pt-4">
        <div className="flex flex-wrap items-center gap-2 text-xs text-zinc-500">
          <span className="hint-chip">Ctrl+Enter analyze</span>
          <span className="hint-chip">Ctrl+K command</span>
          <span className="hint-chip">[ ] tabs</span>
          <span className="hint-chip">J / K bundles</span>
        </div>
        <p className="text-sm text-zinc-500">{busy ? busyDetail || "Background analysis is running." : "Local-first path only. GitHub URLs stay out of scope here."}</p>
      </div>

      {run ? (
        <div className="space-y-2 border-t border-zinc-900 pt-4">
          <div className="flex items-center justify-between gap-3">
            <p className="panel-kicker">Latest run</p>
            <span className="text-xs text-zinc-500">{formatRelativeTime(run.updated_at)}</span>
          </div>
          <div className="space-y-1.5">
            {recentProgress.length ? (
              recentProgress.map((event, index) => (
                <article className="panel-row-static" key={`${event.stage}-${event.status}-${index}`}>
                  <div>
                    <p className="text-[10px] font-semibold uppercase tracking-[0.22em] text-zinc-500">
                      {event.stage} / {event.status}
                    </p>
                    <p className="mt-2 text-sm text-zinc-200">{event.detail || "No additional detail."}</p>
                  </div>
                  <ChevronDown className="size-4 rotate-[-90deg] text-zinc-600" />
                </article>
              ))
            ) : (
              <p className="text-sm text-zinc-500">Waiting for progress events from the backend.</p>
            )}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function splitToLines(value: string) {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
}
