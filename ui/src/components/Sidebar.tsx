import type { ReactNode } from "react";
import {
  Bot,
  BookOpenText,
  FileSearch2,
  FolderPlus,
  FolderCog,
  FolderKanban,
  LayoutDashboard,
  Network,
  PanelRightOpen,
  Shapes,
  Sparkles,
} from "lucide-react";

import type { SavedWorkspace, TabKey } from "@/lib/types";
import { cn, formatRelativeTime } from "@/lib/utils";

interface SidebarProps {
  activeTab: TabKey;
  onCreateWorkspace: () => void;
  onSelectTab: (tab: TabKey) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  workspaces: SavedWorkspace[];
  activeWorkspaceID: string;
}

const viewItems: Array<{ icon: ReactNode; label: string; value: TabKey }> = [
  { icon: <FolderKanban className="size-4" />, label: "Projects", value: "projects" },
  { icon: <FolderCog className="size-4" />, label: "Project", value: "project" },
  { icon: <Bot className="size-4" />, label: "Connections", value: "connections" },
  { icon: <PanelRightOpen className="size-4" />, label: "Properties", value: "properties" },
  { icon: <LayoutDashboard className="size-4" />, label: "Dashboard", value: "dashboard" },
  { icon: <BookOpenText className="size-4" />, label: "Summary", value: "summary" },
  { icon: <Shapes className="size-4" />, label: "Architecture", value: "architecture" },
  { icon: <Network className="size-4" />, label: "Flowchart", value: "flowchart" },
  { icon: <FileSearch2 className="size-4" />, label: "Issues", value: "issues" },
  { icon: <Sparkles className="size-4" />, label: "Recommendations", value: "recommendations" },
];

export function Sidebar({
  activeTab,
  onCreateWorkspace,
  onSelectTab,
  onSelectWorkspace,
  workspaces,
  activeWorkspaceID,
}: SidebarProps) {
  const activeWorkspace = workspaces.find((workspace) => workspace.id === activeWorkspaceID) || null;

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
        <SectionLabel label="Current project" />
        {activeWorkspace ? (
          <button className={cn("sidebar-project", "sidebar-project-active")} onClick={() => onSelectWorkspace(activeWorkspace.id)} type="button">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-zinc-100">{activeWorkspace.label}</p>
                <p className="mt-1 truncate text-xs text-zinc-500">{activeWorkspace.repoPath || "Set a local repository path"}</p>
              </div>
              <span className="text-[10px] text-zinc-600">{formatRelativeTime(activeWorkspace.updatedAt)}</span>
            </div>
          </button>
        ) : (
          <p className="sidebar-empty">No active project yet.</p>
        )}
      </div>
    </aside>
  );
}

function SectionLabel({ label }: { label: string }) {
  return <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-zinc-600">{label}</p>;
}
