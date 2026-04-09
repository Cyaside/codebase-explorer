import type { ReactNode } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import { BookOpenText, FileCode2, Files, FolderClock, Sparkles } from "lucide-react";

import { FlowchartCanvas } from "@/components/FlowchartCanvas";
import { StatsGrid } from "@/components/StatsGrid";
import type { InspectorState, TabKey, WorkbenchBundle } from "@/lib/types";
import { bundleLink, formatRelativeTime } from "@/lib/utils";

interface TabPanelsProps {
  activeTab: TabKey;
  bundle: WorkbenchBundle | null;
  onInspect: (value: InspectorState | null) => void;
  onTabChange: (value: TabKey) => void;
}

export function TabPanels({ activeTab, bundle, onInspect, onTabChange }: TabPanelsProps) {
  return (
    <Tabs.Root className="space-y-4" onValueChange={(value) => onTabChange(value as TabKey)} value={activeTab}>
      <Tabs.List className="tab-strip">
        <TabTrigger value="dashboard">Dashboard</TabTrigger>
        <TabTrigger value="summary">Summary</TabTrigger>
        <TabTrigger value="architecture">Architecture</TabTrigger>
        <TabTrigger value="flowchart">Flowchart</TabTrigger>
        <TabTrigger value="issues">Issues</TabTrigger>
        <TabTrigger value="recommendations">Recommendations</TabTrigger>
      </Tabs.List>

      <Tabs.Content value="dashboard">
        <DashboardView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="summary">
        <SummaryView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="architecture">
        <ArchitectureView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="flowchart">
        <div className="space-y-4">
          <SurfaceHeader
            eyebrow="Flowchart"
            title={bundle ? `${bundle.summary.project_name || bundle.summary.name} graph` : "Project graph"}
            note={
              bundle
                ? "Interactive topology rendered directly in the workbench."
                : "Run an analysis to unlock the project graph."
            }
          />
          <FlowchartCanvas bundle={bundle} onInspect={onInspect} />
        </div>
      </Tabs.Content>

      <Tabs.Content value="issues">
        <IssuesView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="recommendations">
        <RecommendationsView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>
    </Tabs.Root>
  );
}

function DashboardView({ bundle, onInspect }: { bundle: WorkbenchBundle | null; onInspect: (value: InspectorState | null) => void }) {
  if (!bundle) {
    return <EmptyState body="Open a project and run analysis to populate the dashboard." title="No active project yet" />;
  }

  return (
    <div className="space-y-4">
      <SurfaceHeader
        eyebrow="Dashboard"
        title={bundle.summary.project_name || bundle.summary.name}
        note={`${bundle.summary.project_type || "Repository"} · ${bundle.summary.total_files} files · ${bundle.summary.total_lines} lines`}
      />
      <StatsGrid bundle={bundle} />
      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.08fr)_minmax(0,1fr)]">
        <ListPanel
          icon={<FolderClock className="size-4" />}
          title="Reading path"
          rows={bundle.data.reading_path.map((item) => ({
            id: item.path,
            label: item.path,
            meta: item.reason,
            trailing: "Recommended",
            onSelect: () =>
              onInspect({
                eyebrow: "Reading path",
                title: item.path,
                description: item.reason,
                notes: [bundle.data.ai.reading_path_explanations.find((entry) => entry.path === item.path)?.rationale || "No AI rationale recorded."],
                properties: [{ label: "Path", value: item.path }],
              }),
          }))}
        />
        <ListPanel
          icon={<Sparkles className="size-4" />}
          title="Hotspot guidance"
          rows={bundle.data.ai.hotspot_explanations.map((item) => ({
            id: item.path,
            label: item.path,
            meta: item.explanation,
            trailing: formatRelativeTime(bundle.summary.generated_at),
            onSelect: () =>
              onInspect({
                eyebrow: "Hotspot",
                title: item.path,
                description: item.explanation,
                notes: [bundle.data.ai.note || "AI synthesis is available for this bundle."],
                properties: [{ label: "Path", value: item.path }],
              }),
          }))}
        />
      </div>
    </div>
  );
}

function SummaryView({ bundle, onInspect }: { bundle: WorkbenchBundle | null; onInspect: (value: InspectorState | null) => void }) {
  if (!bundle) {
    return <EmptyState body="Summary becomes available once the workspace has a bundle." title="No summary yet" />;
  }

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
      <TextPanel body={bundle.data.project.summary || bundle.data.ai.project_summary || "No summary available."} eyebrow="Summary" title="Project snapshot" />
      <div className="grid gap-4">
        <ListPanel
          icon={<Files className="size-4" />}
          title="Languages"
          rows={bundle.data.languages.map((item) => ({
            id: item.name,
            label: item.name,
            meta: `${item.file_count} files · ${item.line_count} lines`,
            trailing: "",
            onSelect: () =>
              onInspect({
                eyebrow: "Language",
                title: item.name,
                description: "Language footprint in the active project bundle.",
                notes: [`${item.file_count} files`, `${item.line_count} lines`],
                properties: [
                  { label: "Files", value: String(item.file_count) },
                  { label: "Lines", value: String(item.line_count) },
                ],
              }),
          }))}
        />
        <ListPanel
          icon={<BookOpenText className="size-4" />}
          title="Important directories"
          rows={bundle.data.important_directories.map((item) => ({
            id: item,
            label: item,
            meta: "High-signal directory",
            trailing: "",
            onSelect: () =>
              onInspect({
                eyebrow: "Directory",
                title: item,
                description: "Marked as an important directory by the deterministic analyzer.",
                notes: [bundle.summary.analyzed_path || "Local checkout root unavailable."],
                properties: [{ label: "Path", value: item }],
              }),
          }))}
        />
      </div>
    </div>
  );
}

function ArchitectureView({ bundle, onInspect }: { bundle: WorkbenchBundle | null; onInspect: (value: InspectorState | null) => void }) {
  if (!bundle) {
    return <EmptyState body="Architecture details appear after a successful analysis." title="No architecture view yet" />;
  }

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(0,0.95fr)]">
      <TextPanel
        body={bundle.data.ai.architecture_narrative || "AI architecture narrative is unavailable for this bundle."}
        eyebrow="Architecture"
        title="System narrative"
      />
      <ListPanel
        icon={<FileCode2 className="size-4" />}
        title="Module inventory"
        rows={bundle.data.modules.map((item) => ({
          id: item.path,
          label: item.path,
          meta: `${item.file_count} files · ${item.total_lines} lines`,
          trailing: `${item.entry_point_count} entry`,
          onSelect: () =>
            onInspect({
              eyebrow: "Module",
              title: item.path,
              description: "Structured module record from the deterministic analyzer.",
              notes: [bundle.data.ai.hotspot_explanations.find((entry) => entry.path.startsWith(item.path))?.explanation || "No hotspot note recorded."],
              properties: [
                { label: "Files", value: String(item.file_count) },
                { label: "Lines", value: String(item.total_lines) },
                { label: "Entry points", value: String(item.entry_point_count) },
              ],
            }),
        }))}
      />
    </div>
  );
}

function IssuesView({ bundle, onInspect }: { bundle: WorkbenchBundle | null; onInspect: (value: InspectorState | null) => void }) {
  if (!bundle) {
    return <EmptyState body="Link support files to a workspace to unlock issue tracking signals." title="No issue signals yet" />;
  }

  const mentionRows = bundle.data.changes.frequently_mentioned_areas.map((item) => ({
    id: item.path,
    label: item.path,
    meta: `${item.mention_count} mention(s)`,
    trailing: item.confidence,
    onSelect: () =>
      onInspect({
        eyebrow: "Issue signal",
        title: item.path,
        description: "Frequently mentioned across linked support material.",
        notes: [bundle.data.changes.note || "No bundle-level issue note recorded."],
        properties: [
          { label: "Mentions", value: String(item.mention_count) },
          { label: "Confidence", value: item.confidence },
        ],
      }),
  }));

  const sourceRows = bundle.data.changes.sources.map((item) => ({
    id: item.path,
    label: item.path,
    meta: item.kind,
    trailing: item.status,
    onSelect: () =>
      onInspect({
        eyebrow: "Support source",
        title: item.path,
        description: "Source file used to enrich change awareness.",
        notes: [bundle.data.changes.note || "No bundle-level issue note recorded."],
        properties: [
          { label: "Kind", value: item.kind },
          { label: "Status", value: item.status },
        ],
      }),
  }));

  const themeRows = bundle.data.changes.repeated_themes.map((item) => ({
    id: item.name,
    label: item.name,
    meta: "Repeated theme",
    trailing: `${item.mention_count}`,
    onSelect: () =>
      onInspect({
        eyebrow: "Theme",
        title: item.name,
        description: "Repeated theme extracted from linked support files.",
        notes: [`Mention count: ${item.mention_count}`],
        properties: [{ label: "Mentions", value: String(item.mention_count) }],
      }),
  }));

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,0.9fr)]">
      <ListPanel icon={<Sparkles className="size-4" />} title="Mentioned areas" rows={mentionRows} />
      <ListPanel icon={<Files className="size-4" />} title="Support sources" rows={sourceRows} />
      <ListPanel icon={<BookOpenText className="size-4" />} title="Repeated themes" rows={themeRows} />
    </div>
  );
}

function RecommendationsView({ bundle, onInspect }: { bundle: WorkbenchBundle | null; onInspect: (value: InspectorState | null) => void }) {
  if (!bundle) {
    return <EmptyState body="Recommendations appear when the workspace has a current bundle." title="No recommendations yet" />;
  }

  const nextMoves = [
    { label: "Open bundle README", href: bundleLink(bundle.summary.name, "README.md") },
    { label: "Open static viewer", href: bundleLink(bundle.summary.name, "ui/index.html") },
    { label: "Open architecture artifact", href: bundleLink(bundle.summary.name, bundle.data.links.architecture_diagram) },
  ];

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
      <ListPanel
        icon={<BookOpenText className="size-4" />}
        title="Reading path"
        rows={bundle.data.reading_path.map((item) => ({
          id: item.path,
          label: item.path,
          meta: item.reason,
          trailing: "Read",
          onSelect: () =>
            onInspect({
              eyebrow: "Recommendation",
              title: item.path,
              description: item.reason,
              notes: [bundle.data.ai.reading_path_explanations.find((entry) => entry.path === item.path)?.rationale || "No AI rationale recorded."],
              properties: [{ label: "Path", value: item.path }],
            }),
        }))}
      />
      <section className="panel-block space-y-4">
        <div>
          <p className="panel-kicker">Next moves</p>
          <h3 className="mt-3 text-lg font-semibold text-zinc-100">Useful follow-up actions</h3>
        </div>
        <div className="space-y-2">
          {nextMoves.map((item) => (
            <a className="panel-link-row" href={item.href} key={item.label} rel="noreferrer" target="_blank">
              <span>{item.label}</span>
              <span className="text-zinc-500">open</span>
            </a>
          ))}
        </div>
      </section>
    </div>
  );
}

function TabTrigger({ children, value }: { children: ReactNode; value: TabKey }) {
  return (
    <Tabs.Trigger className="tab-trigger" value={value}>
      {children}
    </Tabs.Trigger>
  );
}

function SurfaceHeader({ eyebrow, title, note }: { eyebrow: string; title: string; note: string }) {
  return (
    <section className="panel-block">
      <p className="panel-kicker">{eyebrow}</p>
      <h2 className="mt-3 text-2xl font-semibold tracking-tight text-zinc-50">{title}</h2>
      <p className="mt-3 text-sm leading-6 text-zinc-400">{note}</p>
    </section>
  );
}

function TextPanel({ body, eyebrow, title }: { body: string; eyebrow: string; title: string }) {
  return (
    <section className="panel-block">
      <p className="panel-kicker">{eyebrow}</p>
      <h3 className="mt-3 text-lg font-semibold text-zinc-100">{title}</h3>
      <p className="mt-4 text-sm leading-7 text-zinc-300">{body}</p>
    </section>
  );
}

function ListPanel({
  icon,
  rows,
  title,
}: {
  icon: ReactNode;
  rows: Array<{ id: string; label: string; meta: string; trailing: string; onSelect: () => void }>;
  title: string;
}) {
  return (
    <section className="panel-block">
      <div className="flex items-center gap-3">
        <span className="text-zinc-500">{icon}</span>
        <p className="panel-kicker">{title}</p>
      </div>
      <div className="mt-4 space-y-1.5">
        {rows.length ? (
          rows.map((row) => (
            <button className="panel-row" key={row.id} onClick={row.onSelect} type="button">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-zinc-100">{row.label}</p>
                <p className="mt-1 line-clamp-2 text-xs leading-5 text-zinc-500">{row.meta}</p>
              </div>
              <span className="shrink-0 text-xs text-zinc-500">{row.trailing}</span>
            </button>
          ))
        ) : (
          <p className="text-sm text-zinc-500">Nothing to show yet.</p>
        )}
      </div>
    </section>
  );
}

function EmptyState({ body, title }: { body: string; title: string }) {
  return (
    <section className="panel-block flex min-h-[18rem] items-center justify-center">
      <div className="max-w-md text-center">
        <p className="panel-kicker">Empty state</p>
        <h3 className="mt-3 text-lg font-semibold text-zinc-100">{title}</h3>
        <p className="mt-3 text-sm leading-6 text-zinc-500">{body}</p>
      </div>
    </section>
  );
}
