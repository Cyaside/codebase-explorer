import type { ReactNode } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import { BookOpenText, Files, ShieldAlert, Sparkles } from "lucide-react";

import type { TabKey, WorkbenchBundle } from "@/lib/types";
import { bundleLink } from "@/lib/utils";

interface TabPanelsProps {
  activeTab: TabKey;
  bundle: WorkbenchBundle | null;
  onTabChange: (value: TabKey) => void;
}

export function TabPanels({ activeTab, bundle, onTabChange }: TabPanelsProps) {
  return (
    <Tabs.Root className="space-y-4" onValueChange={(value) => onTabChange(value as TabKey)} value={activeTab}>
      <Tabs.List className="grid gap-2 rounded-2xl border border-border bg-card/80 p-2 md:grid-cols-5">
        <TabTrigger value="summary">Summary</TabTrigger>
        <TabTrigger value="architecture">Architecture</TabTrigger>
        <TabTrigger value="flowchart">Flowchart</TabTrigger>
        <TabTrigger value="issues">Issue Tracking</TabTrigger>
        <TabTrigger value="recommendations">Recommendations</TabTrigger>
      </Tabs.List>

      <Tabs.Content value="summary">
        <PanelLayout
          bundle={bundle}
          cards={[
            ["Summary", <p>{bundle?.data.project.summary || bundle?.data.ai.project_summary || "No summary available."}</p>],
            ["Warnings", <TextList emptyText="No runtime warnings recorded." items={bundle?.data.warnings || []} />],
            [
              "Languages",
              <TextList
                emptyText="No language data."
                items={(bundle?.data.languages || []).map((item) => `${item.name} · ${item.file_count} files · ${item.line_count} lines`)}
              />,
            ],
            ["Entry points", <TextList emptyText="No entry points recorded." items={bundle?.data.entry_points || []} />],
            ["Important directories", <TextList emptyText="No important directories recorded." items={bundle?.data.important_directories || []} />],
          ]}
          emptyMessage="Run an analysis or pick a bundle from the sidebar."
          title="Project snapshot"
        />
      </Tabs.Content>

      <Tabs.Content value="architecture">
        <PanelLayout
          bundle={bundle}
          cards={[
            ["Narrative", <p>{bundle?.data.ai.architecture_narrative || "AI architecture narrative is unavailable for this bundle."}</p>],
            ["Core modules", <TextList emptyText="No core module list recorded." items={bundle?.data.core_modules || []} />],
            [
              "Module inventory",
              <TextList
                emptyText="No module inventory recorded."
                items={(bundle?.data.modules || []).slice(0, 12).map((item) => `${item.path} · ${item.file_count} files · ${item.total_lines} lines`)}
              />,
            ],
          ]}
          emptyMessage="Architecture details will appear after a bundle is selected."
          title="Architecture"
        />
      </Tabs.Content>

      <Tabs.Content value="flowchart">
        <PanelLayout
          bundle={bundle}
          cards={[
            [
              "Module graph",
              <div className="space-y-3">
                {bundle ? (
                  <a
                    className="inline-flex text-sm text-primary transition hover:text-primary-foreground"
                    href={bundleLink(bundle.summary.name, bundle.data.links.architecture_diagram)}
                    rel="noreferrer"
                    target="_blank"
                  >
                    Open raw Mermaid file
                  </a>
                ) : null}
                <pre className="code-block">{bundle?.data.mermaid.architecture || "No architecture diagram source."}</pre>
              </div>,
            ],
            [
              "Dependency graph",
              <div className="space-y-3">
                {bundle ? (
                  <a
                    className="inline-flex text-sm text-primary transition hover:text-primary-foreground"
                    href={bundleLink(bundle.summary.name, bundle.data.links.dependency_diagram)}
                    rel="noreferrer"
                    target="_blank"
                  >
                    Open raw Mermaid file
                  </a>
                ) : null}
                <pre className="code-block">{bundle?.data.mermaid.dependencies || "No dependency diagram source."}</pre>
              </div>,
            ],
            [
              "Top modules",
              <TextList
                emptyText="No module topology available."
                items={(bundle?.data.modules || []).slice(0, 8).map((item) => `${item.path} · ${item.entry_point_count} entry point(s)`)}
              />,
            ],
          ]}
          emptyMessage="Flowchart views depend on a selected bundle."
          title="Codebase flowchart"
        />
      </Tabs.Content>

      <Tabs.Content value="issues">
        <PanelLayout
          bundle={bundle}
          cards={[
            ["Change note", <p>{bundle?.data.changes.note || "No support files were linked to this bundle."}</p>],
            [
              "Support files",
              <TextList
                emptyText="No support files were recorded."
                items={(bundle?.data.changes.sources || []).map((item) => `${item.path} · ${item.kind} · ${item.status}`)}
              />,
            ],
            [
              "Frequently mentioned areas",
              <TextList
                emptyText="No repository areas were matched strongly enough."
                items={(bundle?.data.changes.frequently_mentioned_areas || []).map(
                  (item) => `${item.path} · ${item.mention_count} mention(s) · ${item.confidence}`,
                )}
              />,
            ],
            [
              "Repeated themes",
              <TextList
                emptyText="No repeated themes crossed the reporting threshold."
                items={(bundle?.data.changes.repeated_themes || []).map((item) => `${item.name} · ${item.mention_count} mention(s)`)}
              />,
            ],
          ]}
          emptyMessage="Issue tracking becomes available when a bundle is selected."
          title="Issue tracking"
        />
      </Tabs.Content>

      <Tabs.Content value="recommendations">
        <PanelLayout
          bundle={bundle}
          cards={[
            [
              "Reading path",
              <TextList
                emptyText="No reading path recommendation available."
                items={
                  (bundle?.data.ai.reading_path_explanations || []).length
                    ? (bundle?.data.ai.reading_path_explanations || []).map((item) => `${item.path} · ${item.rationale}`)
                    : (bundle?.data.reading_path || []).map((item) => `${item.path} · ${item.reason}`)
                }
              />,
            ],
            [
              "Hotspot guidance",
              <TextList
                emptyText="AI hotspot notes are unavailable for this bundle."
                items={(bundle?.data.ai.hotspot_explanations || []).map((item) => `${item.path} · ${item.explanation}`)}
              />,
            ],
            [
              "Next moves",
              <TextList
                emptyText="No follow-up recommendations available."
                items={
                  bundle
                    ? [
                        `Open raw README · ${bundleLink(bundle.summary.name, "README.md")}`,
                        `Open viewer snapshot · ${bundleLink(bundle.summary.name, "ui/index.html")}`,
                        `Review bundle warnings · ${bundle.data.warnings.length} warning(s)`,
                      ]
                    : []
                }
              />,
            ],
          ]}
          emptyMessage="Recommendations will appear after selecting a bundle."
          title="Recommendations"
        />
      </Tabs.Content>
    </Tabs.Root>
  );
}

function TabTrigger({ children, value }: { children: ReactNode; value: TabKey }) {
  return (
    <Tabs.Trigger
      className="rounded-xl px-3 py-2 text-sm font-medium text-muted-foreground transition data-[state=active]:bg-accent data-[state=active]:text-accent-foreground"
      value={value}
    >
      {children}
    </Tabs.Trigger>
  );
}

function PanelLayout({
  bundle,
  cards,
  emptyMessage,
  title,
}: {
  bundle: WorkbenchBundle | null;
  cards: Array<[string, ReactNode]>;
  emptyMessage: string;
  title: string;
}) {
  if (!bundle) {
    return <div className="rounded-3xl border border-dashed border-border bg-card/50 p-10 text-sm text-muted-foreground">{emptyMessage}</div>;
  }

  return (
    <div className="space-y-4">
      <div className="rounded-3xl border border-border bg-card/90 p-5 shadow-sm">
        <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
          <div>
            <p className="text-[10px] uppercase tracking-[0.28em] text-primary">{title}</p>
            <h3 className="mt-2 text-xl font-bold">{bundle.summary.project_name || bundle.summary.name}</h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {bundle.summary.project_type || "Repository"} · {bundle.summary.total_files} files · {bundle.summary.total_lines} lines
            </p>
          </div>

          <div className="flex flex-wrap gap-2">
            <QuickLink href={bundleLink(bundle.summary.name, "README.md")} label="README" />
            <QuickLink href={bundleLink(bundle.summary.name, "ui/index.html")} label="Viewer" />
          </div>
        </div>

        <div className="mt-5 grid gap-3 lg:grid-cols-2 2xl:grid-cols-4">
          <SignalChip
            icon={<Sparkles className="size-4 text-primary" />}
            label="AI"
            note={bundle.data.ai.provider || "deterministic"}
            value={bundle.data.ai.status || "disabled"}
          />
          <SignalChip
            icon={<ShieldAlert className="size-4 text-primary" />}
            label="Warnings"
            note="Runtime checks"
            value={String(bundle.data.warnings.length)}
          />
          <SignalChip
            icon={<Files className="size-4 text-primary" />}
            label="Support files"
            note="Change awareness inputs"
            value={String(bundle.data.changes.support_file_count)}
          />
          <SignalChip
            icon={<BookOpenText className="size-4 text-primary" />}
            label="Reading path"
            note="Suggested checkpoints"
            value={String(bundle.data.reading_path.length)}
          />
        </div>
      </div>

      <div className="grid gap-4 xl:grid-cols-2">
        {cards.map(([titleText, body]) => (
          <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-sm" key={titleText}>
            <p className="text-[10px] uppercase tracking-[0.28em] text-muted-foreground">{titleText}</p>
            <div className="mt-4 text-sm leading-6 text-foreground">{body}</div>
          </section>
        ))}
      </div>
    </div>
  );
}

function TextList({ emptyText, items }: { emptyText: string; items: string[] }) {
  if (!items.length) {
    return <p className="text-sm text-muted-foreground">{emptyText}</p>;
  }

  return (
    <ul className="space-y-2">
      {items.map((item, index) => (
        <li className="rounded-2xl border border-border bg-background/60 px-3 py-3 text-sm text-muted-foreground" key={`${item}-${index}`}>
          <div className="flex gap-3">
            <span className="mt-0.5 text-xs font-semibold uppercase tracking-[0.18em] text-primary/75">
              {String(index + 1).padStart(2, "0")}
            </span>
            <span>{item}</span>
          </div>
        </li>
      ))}
    </ul>
  );
}

function SignalChip({
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
    <article className="rounded-2xl border border-border bg-background/55 px-3 py-3">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{label}</span>
        {icon}
      </div>
      <strong className="mt-3 block text-base font-semibold text-foreground">{value}</strong>
      <p className="mt-1 text-xs text-muted-foreground">{note}</p>
    </article>
  );
}

function QuickLink({ href, label }: { href: string; label: string }) {
  return (
    <a
      className="inline-flex items-center rounded-full border border-border bg-background/60 px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition hover:border-border-strong hover:bg-background hover:text-foreground"
      href={href}
      rel="noreferrer"
      target="_blank"
    >
      {label}
    </a>
  );
}
