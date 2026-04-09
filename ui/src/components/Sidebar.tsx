import type { ReactNode } from "react";
import {
  Bot,
  BookOpenText,
  FileSearch2,
  FolderPlus,
  LayoutDashboard,
  Network,
  Shapes,
  Sparkles,
} from "lucide-react";

import type { BundleSummary, ConnectionProfile, SavedWorkspace, TabKey } from "@/lib/types";
import { cn, formatRelativeTime } from "@/lib/utils";

interface SidebarProps {
  activeTab: TabKey;
  bundles: BundleSummary[];
  onCreateWorkspace: () => void;
  onSelectBundle: (bundleName: string) => void;
  onSelectProfile: (profileID: string) => void;
  onSelectTab: (tab: TabKey) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  profiles: ConnectionProfile[];
  selectedBundle: string;
  selectedProfile: string;
  workspaceBundles: BundleSummary[];
  workspaces: SavedWorkspace[];
  activeWorkspaceID: string;
}

const viewItems: Array<{ icon: ReactNode; label: string; value: TabKey }> = [
  { icon: <LayoutDashboard className="size-4" />, label: "Dashboard", value: "dashboard" },
  { icon: <BookOpenText className="size-4" />, label: "Summary", value: "summary" },
  { icon: <Shapes className="size-4" />, label: "Architecture", value: "architecture" },
  { icon: <Network className="size-4" />, label: "Flowchart", value: "flowchart" },
  { icon: <FileSearch2 className="size-4" />, label: "Issues", value: "issues" },
  { icon: <Sparkles className="size-4" />, label: "Recommendations", value: "recommendations" },
];

export function Sidebar({
  activeTab,
  bundles,
  onCreateWorkspace,
  onSelectBundle,
  onSelectProfile,
  onSelectTab,
  onSelectWorkspace,
  profiles,
  selectedBundle,
  selectedProfile,
  workspaceBundles,
  workspaces,
  activeWorkspaceID,
}: SidebarProps) {
  return (
    <aside className="sidebar-shell">
      <div className="space-y-5">
        <div className="flex items-center gap-3">
          <div className="grid size-10 place-items-center rounded-2xl bg-zinc-100 text-sm font-bold text-black">C</div>
          <div>
            <p className="text-lg font-semibold text-zinc-50">Codebase Explorer</p>
            <p className="text-xs text-zinc-500">Local control plane</p>
          </div>
        </div>

        <button className="sidebar-primary" onClick={onCreateWorkspace} type="button">
          <FolderPlus className="size-4" />
          New Project
        </button>
      </div>

      <div className="sidebar-section">
        <SectionLabel label="Views" />
        <div className="space-y-1">
          {viewItems.map((item) => (
            <button
              className={cn("sidebar-item", activeTab === item.value && "sidebar-item-active")}
              key={item.value}
              onClick={() => onSelectTab(item.value)}
              type="button"
            >
              {item.icon}
              <span>{item.label}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="sidebar-section">
        <SectionLabel label="Projects" />
        <div className="space-y-1.5">
          {workspaces.length ? (
            workspaces.map((workspace) => (
              <button
                className={cn("sidebar-project", workspace.id === activeWorkspaceID && "sidebar-project-active")}
                key={workspace.id}
                onClick={() => onSelectWorkspace(workspace.id)}
                type="button"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-zinc-100">{workspace.label}</p>
                    <p className="mt-1 truncate text-xs text-zinc-500">{workspace.repoPath || "Set a local repository path"}</p>
                  </div>
                  <span className="text-[10px] text-zinc-600">{formatRelativeTime(workspace.updatedAt)}</span>
                </div>
              </button>
            ))
          ) : (
            <p className="sidebar-empty">No saved project yet.</p>
          )}
        </div>
      </div>

      <div className="sidebar-section">
        <SectionLabel label="Connections" />
        <div className="space-y-1.5">
          {profiles.map((profile) => (
            <button
              className={cn("sidebar-item", profile.id === selectedProfile && "sidebar-item-active")}
              key={profile.id}
              onClick={() => onSelectProfile(profile.id)}
              type="button"
            >
              <Bot className="size-4" />
              <div className="min-w-0 text-left">
                <p className="truncate text-sm">{profile.label}</p>
                <p className="truncate text-xs text-zinc-500">{profile.provider?.name || "deterministic only"}</p>
              </div>
            </button>
          ))}
        </div>
      </div>

      <div className="sidebar-section">
        <SectionLabel label="Bundles" />
        <div className="space-y-1.5">
          {(workspaceBundles.length ? workspaceBundles : bundles.slice(0, 6)).map((bundle) => (
            <button
              className={cn("sidebar-project", bundle.name === selectedBundle && "sidebar-project-active")}
              key={bundle.name}
              onClick={() => onSelectBundle(bundle.name)}
              type="button"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-zinc-100">{bundle.project_name || bundle.name}</p>
                  <p className="mt-1 truncate text-xs text-zinc-500">
                    {bundle.ai_status || "deterministic"} · {bundle.total_files} files
                  </p>
                </div>
                <span className="text-[10px] text-zinc-600">{formatRelativeTime(bundle.generated_at)}</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    </aside>
  );
}

function SectionLabel({ label }: { label: string }) {
  return <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-zinc-600">{label}</p>;
}
