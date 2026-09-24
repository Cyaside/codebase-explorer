import type { ReactNode } from "react";
import {
  ArrowUpRight, BookOpenText, CircleHelp, FileSearch2, FolderCog, FolderKanban,
  FolderPlus, GitBranch, LayoutDashboard, Network, PanelRightOpen, Shapes, Waypoints,
} from "lucide-react";

import type { SavedWorkspace, TabKey } from "@/lib/types";
import { cn } from "@/lib/utils";

interface SidebarProps {
  activeTab: TabKey;
  onCreateWorkspace: () => void;
  onSelectTab: (tab: TabKey) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  workspaces: SavedWorkspace[];
  activeWorkspaceID: string;
}

const workspaceItems: Array<{ icon: ReactNode; label: string; value: TabKey }> = [
  { icon: <FolderKanban />, label: "Projects", value: "projects" },
  { icon: <FolderCog />, label: "Project setup", value: "project" },
  { icon: <Waypoints />, label: "Connections", value: "connections" },
];

const analysisItems: Array<{ icon: ReactNode; label: string; value: TabKey }> = [
  { icon: <LayoutDashboard />, label: "Overview", value: "dashboard" },
  { icon: <BookOpenText />, label: "Summary", value: "summary" },
  { icon: <Shapes />, label: "Architecture", value: "architecture" },
  { icon: <Network />, label: "Graphs", value: "flowchart" },
  { icon: <FileSearch2 />, label: "Issues", value: "issues" },
  { icon: <GitBranch />, label: "Recommendations", value: "recommendations" },
  { icon: <PanelRightOpen />, label: "Inspector", value: "properties" },
];

export function Sidebar({ activeTab, onCreateWorkspace, onSelectTab, onSelectWorkspace, workspaces, activeWorkspaceID }: SidebarProps) {
  const activeWorkspace = workspaces.find((workspace) => workspace.id === activeWorkspaceID) || null;
  const renderItem = (item: { icon: ReactNode; label: string; value: TabKey }) => (
    <button
      aria-current={activeTab === item.value ? "page" : undefined}
      className={cn("sidebar-item", activeTab === item.value && "sidebar-item-active")}
      key={item.value}
      onClick={() => onSelectTab(item.value)}
      type="button"
    >
      <span className="sidebar-item-icon">{item.icon}</span>
      <span>{item.label}</span>
    </button>
  );

  return (
    <aside className="sidebar-shell" aria-label="Workbench navigation">
      <div className="sidebar-brand">
        <img className="brand-logo" src="/assets/brand-mark.svg" alt="" width="31" height="31" />
        <div className="min-w-0">
          <strong>Codebase Explorer</strong>
          <small>Repository intelligence</small>
        </div>
      </div>
      <nav aria-label="Primary" className="sidebar-navigation">
        <div className="sidebar-section">
          <SectionLabel label="Workspace" />
          <div className="sidebar-items">{workspaceItems.map(renderItem)}</div>
        </div>
        <div className="sidebar-section">
          <SectionLabel label="Analysis" />
          <div className="sidebar-items">{analysisItems.map(renderItem)}</div>
        </div>
      </nav>
      <div className="sidebar-bottom">
        <button className="sidebar-primary" onClick={onCreateWorkspace} type="button">
          <FolderPlus className="size-4" /> New project <ArrowUpRight className="ml-auto size-3.5" />
        </button>
        {activeWorkspace ? (
          <button className="sidebar-project" onClick={() => onSelectWorkspace(activeWorkspace.id)} type="button">
            <span className="sidebar-project-icon"><PanelRightOpen className="size-4" /></span>
            <span className="min-w-0 flex-1 text-left">
              <strong className="block truncate">{activeWorkspace.label}</strong>
              <small className="block truncate">{activeWorkspace.repoPath || "Repository path needed"}</small>
            </span>
          </button>
        ) : (
          <div className="sidebar-project"><CircleHelp className="size-4" /><span>No project selected</span></div>
        )}
      </div>
    </aside>
  );
}

function SectionLabel({ label }: { label: string }) {
  return <p className="sidebar-label">{label}</p>;
}
