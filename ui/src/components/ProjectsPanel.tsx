import { FolderPlus, Layers3, Trash2 } from "lucide-react";

import type { BundleSummary, SavedWorkspace } from "@/lib/types";
import { cn, formatRelativeTime, normalizeLocalPath } from "@/lib/utils";

interface ProjectsPanelProps {
  activeWorkspaceID: string;
  bundles: BundleSummary[];
  onCreateWorkspace: () => void;
  onDeleteBundle: (bundleName: string) => void;
  onDeleteWorkspace: (workspaceID: string) => void;
  onSelectBundle: (bundleName: string, workspaceID: string) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  selectedBundle: string;
  workspaces: SavedWorkspace[];
}

export function ProjectsPanel({
  activeWorkspaceID,
  bundles,
  onCreateWorkspace,
  onDeleteBundle,
  onDeleteWorkspace,
  onSelectBundle,
  onSelectWorkspace,
  selectedBundle,
  workspaces,
}: ProjectsPanelProps) {
  const activeWorkspace = workspaces.find((workspace) => workspace.id === activeWorkspaceID) || null;
  const relatedBundles = activeWorkspace ? bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, activeWorkspace.repoPath)) : [];

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
      <section className="panel-block space-y-4">
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className="panel-kicker">Projects</p>
            <h3 className="mt-3 text-lg font-semibold text-zinc-100">Saved workspaces</h3>
          </div>
          <button className="secondary-control" onClick={onCreateWorkspace} type="button">
            <FolderPlus className="size-4" />
            Add
          </button>
        </div>

        <div className="space-y-2">
          {workspaces.length ? (
            workspaces.map((workspace) => {
              const workspaceBundles = bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, workspace.repoPath));
              return (
                <article
                  className={cn("panel-row-static", workspace.id === activeWorkspaceID && "border-zinc-700 bg-zinc-950")}
                  key={workspace.id}
                >
                  <button className="min-w-0 flex-1 text-left" onClick={() => onSelectWorkspace(workspace.id)} type="button">
                    <p className="truncate text-sm font-medium text-zinc-100">{workspace.label}</p>
                    <p className="mt-1 truncate text-xs text-zinc-500">{workspace.repoPath || "Set a local repository path."}</p>
                    <div className="mt-3 flex flex-wrap gap-2 text-[11px] text-zinc-600">
                      <span>{workspaceBundles.length} bundle(s)</span>
                      <span>{formatRelativeTime(workspace.updatedAt)}</span>
                    </div>
                  </button>
                  <button
                    aria-label={`Delete ${workspace.label}`}
                    className="secondary-control danger px-3 py-2"
                    onClick={() => onDeleteWorkspace(workspace.id)}
                    type="button"
                  >
                    <Trash2 className="size-4" />
                  </button>
                </article>
              );
            })
          ) : (
            <p className="sidebar-empty">No saved project yet.</p>
          )}
        </div>
      </section>

      <section className="panel-block space-y-4">
        <div className="flex items-center gap-3">
          <Layers3 className="size-4 text-zinc-500" />
          <div>
            <p className="panel-kicker">Project bundles</p>
            <h3 className="mt-3 text-lg font-semibold text-zinc-100">
              {activeWorkspace ? activeWorkspace.label : "Select a project"}
            </h3>
          </div>
        </div>

        {activeWorkspace ? (
          relatedBundles.length ? (
            <div className="space-y-2">
              {relatedBundles.map((bundle) => (
                <article
                  className={cn("panel-row-static", bundle.name === selectedBundle && "border-zinc-700 bg-zinc-950")}
                  key={bundle.name}
                >
                  <button className="min-w-0 flex-1 text-left" onClick={() => onSelectBundle(bundle.name, activeWorkspace.id)} type="button">
                    <p className="truncate text-sm font-medium text-zinc-100">{bundle.project_name || bundle.name}</p>
                    <p className="mt-1 truncate text-xs text-zinc-500">
                      {bundle.ai_status || "ready"} · {bundle.total_files} files · {formatRelativeTime(bundle.generated_at)}
                    </p>
                  </button>
                  <button
                    aria-label={`Delete bundle ${bundle.name}`}
                    className="secondary-control danger px-3 py-2"
                    onClick={() => onDeleteBundle(bundle.name)}
                    type="button"
                  >
                    <Trash2 className="size-4" />
                  </button>
                </article>
              ))}
            </div>
          ) : (
            <p className="sidebar-empty">No bundle yet for this project. Run analysis from Project Setup.</p>
          )
        ) : (
          <p className="sidebar-empty">Choose a project from the left column to manage its bundles.</p>
        )}
      </section>
    </div>
  );
}

function matchesWorkspace(bundlePath: string, workspacePath: string) {
  if (!workspacePath.trim()) {
    return false;
  }

  return normalizeLocalPath(bundlePath) === normalizeLocalPath(workspacePath);
}
