import type { ReactNode } from "react";
import { Activity, FolderOpenDot, Info, Sparkles } from "lucide-react";

import { fullAIStatusLabel } from "@/lib/fullAi";
import type { AnalyzeRun, InspectorState, SavedWorkspace, WorkbenchBundle } from "@/lib/types";
import { formatTimestamp } from "@/lib/utils";

interface PropertiesPanelProps {
  activeRun: AnalyzeRun | null;
  bundle: WorkbenchBundle | null;
  inspector: InspectorState | null;
  workspace: SavedWorkspace | null;
}

export function PropertiesPanel({ activeRun, bundle, inspector, workspace }: PropertiesPanelProps) {
  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
      <div className="space-y-4">
        <RailPanel icon={<FolderOpenDot className="size-4" />} title="Workspace">
          {workspace ? (
            <div className="space-y-3 text-sm">
              <PropertyRow label="Label" value={workspace.label} />
              <PropertyRow label="Repository" value={workspace.repoPath || "Not configured"} />
              <PropertyRow label="Support files" value={String(workspace.supportFiles.length)} />
              <PropertyRow label="Ignore patterns" value={String(workspace.ignorePatterns.length)} />
              <PropertyRow label="Updated" value={formatTimestamp(workspace.updatedAt)} />
            </div>
          ) : (
            <p className="text-sm leading-6 text-zinc-500">Create a project to anchor this shell to one local repository.</p>
          )}
        </RailPanel>

        <RailPanel icon={<Activity className="size-4" />} title="Run state">
          {activeRun ? (
            <div className="space-y-3">
              <PropertyRow label="Status" value={activeRun.status} />
              <PropertyRow label="Events" value={String(activeRun.progress.length)} />
              <PropertyRow label="Updated" value={formatTimestamp(activeRun.updated_at)} />
              <p className="text-sm leading-6 text-zinc-400">{activeRun.progress[activeRun.progress.length - 1]?.detail || activeRun.error || "No run details."}</p>
            </div>
          ) : (
            <p className="text-sm leading-6 text-zinc-500">No active run. Start an analyze pass from the project setup tab.</p>
          )}
        </RailPanel>

        <RailPanel icon={<Sparkles className="size-4" />} title="Bundle">
          {bundle ? (
            <div className="space-y-3">
              <PropertyRow label="Bundle" value={bundle.summary.name} />
              <PropertyRow label="Generated" value={formatTimestamp(bundle.summary.generated_at)} />
              <PropertyRow label="AI status" value={bundle.summary.ai_status || "disabled"} />
              <PropertyRow label="Full-AI status" value={fullAIStatusLabel(bundle)} />
              <PropertyRow label="AI functions" value={`${bundle.data.full_ai_execution.executed_count}/${bundle.data.full_ai.prepared_functions}`} />
              <PropertyRow label="Verified" value={String(bundle.data.full_ai_execution.verified_count)} />
              <PropertyRow label="Warnings" value={String(bundle.data.warnings.length)} />
            </div>
          ) : (
            <p className="text-sm leading-6 text-zinc-500">No bundle selected for this workspace yet.</p>
          )}
        </RailPanel>
      </div>

      <RailPanel icon={<Info className="size-4" />} title="Selection">
        {inspector ? (
          <div className="space-y-4">
            <div>
              <p className="panel-kicker">{inspector.eyebrow}</p>
              <h3 className="mt-3 text-lg font-semibold text-zinc-100">{inspector.title}</h3>
              <p className="mt-3 text-sm leading-6 text-zinc-400">{inspector.description}</p>
            </div>

            {inspector.properties.length ? (
              <div className="space-y-2 border-t border-zinc-900 pt-4">
                {inspector.properties.map((item) => (
                  <PropertyRow key={`${item.label}-${item.value}`} label={item.label} value={item.value} />
                ))}
              </div>
            ) : null}

            {inspector.notes.length ? (
              <div className="space-y-2 border-t border-zinc-900 pt-4">
                {inspector.notes.map((note, index) => (
                  <p className="text-sm leading-6 text-zinc-500" key={`${note}-${index}`}>
                    {note}
                  </p>
                ))}
              </div>
            ) : null}
          </div>
        ) : (
          <p className="text-sm leading-6 text-zinc-500">Select a graph node or an issue signal and this tab will show its details.</p>
        )}
      </RailPanel>
    </div>
  );
}

function RailPanel({ children, icon, title }: { children: ReactNode; icon: ReactNode; title: string }) {
  return (
    <section className="rail-section">
      <div className="flex items-center gap-3">
        <span className="text-zinc-500">{icon}</span>
        <p className="panel-kicker">{title}</p>
      </div>
      <div className="mt-4">{children}</div>
    </section>
  );
}

function PropertyRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-3 text-sm">
      <span className="text-zinc-500">{label}</span>
      <span className="max-w-[20rem] text-right text-zinc-100">{value}</span>
    </div>
  );
}
