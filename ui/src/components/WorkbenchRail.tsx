import type { ReactNode } from "react";
import { Activity, Bot, FileWarning, FolderTree, GitBranch, Link2, Sparkles } from "lucide-react";

import type { AnalyzeRun, WorkbenchBundle } from "@/lib/types";
import { bundleLink } from "@/lib/utils";

interface WorkbenchRailProps {
  activeRun: AnalyzeRun | null;
  bundle: WorkbenchBundle | null;
}

export function WorkbenchRail({ activeRun, bundle }: WorkbenchRailProps) {
  return (
    <aside className="space-y-4 2xl:sticky 2xl:top-5 2xl:max-h-[calc(100vh-2.5rem)] 2xl:overflow-y-auto">
      <RailSection
        icon={<Activity className="size-4 text-primary" />}
        title="Run state"
        body={<RunStateCard run={activeRun} />}
      />

      <RailSection
        icon={<Sparkles className="size-4 text-primary" />}
        title="Bundle health"
        body={<BundleHealthCard bundle={bundle} />}
      />

      <RailSection
        icon={<FolderTree className="size-4 text-primary" />}
        title="Signals"
        body={<BundleSignalsCard bundle={bundle} />}
      />

      <RailSection
        icon={<Link2 className="size-4 text-primary" />}
        title="Quick links"
        body={<QuickLinksCard bundle={bundle} />}
      />
    </aside>
  );
}

function RailSection({ body, icon, title }: { body: ReactNode; icon: ReactNode; title: string }) {
  return (
    <section className="rounded-3xl border border-border bg-card/90 p-4 shadow-sm">
      <div className="mb-4 flex items-center gap-2">
        {icon}
        <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">{title}</p>
      </div>
      {body}
    </section>
  );
}

function RunStateCard({ run }: { run: AnalyzeRun | null }) {
  if (!run) {
    return <p className="text-sm text-muted-foreground">No analyze run yet. Start a run to see stage progress and terminal state here.</p>;
  }

  const latestEvent = run.progress[run.progress.length - 1];
  return (
    <div className="space-y-4">
      <div className="rounded-2xl border border-border bg-background/60 p-3">
        <div className="flex items-center justify-between gap-3">
          <p className="text-sm font-semibold text-foreground">Run {run.id}</p>
          <span className={runStatusPillClassName(run.status)}>{run.status}</span>
        </div>
        <p className="mt-3 text-sm text-muted-foreground">{latestEvent?.detail || "Waiting for backend progress."}</p>
      </div>

      <dl className="space-y-3">
        <PropertyRow label="Latest stage" value={latestEvent?.stage || "analyze"} />
        <PropertyRow label="Progress events" value={String(run.progress.length)} />
        <PropertyRow label="Updated" value={formatTimestamp(run.updated_at)} />
      </dl>
    </div>
  );
}

function BundleHealthCard({ bundle }: { bundle: WorkbenchBundle | null }) {
  if (!bundle) {
    return <p className="text-sm text-muted-foreground">Select a bundle to inspect AI status, warnings, and change-awareness coverage.</p>;
  }

  const warningCount = bundle.data.warnings.length;
  const supportFileCount = bundle.data.changes.support_file_count;
  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 2xl:grid-cols-1">
        <MiniMetric
          icon={<Bot className="size-4 text-primary" />}
          label="AI status"
          note={bundle.data.ai.provider || "deterministic"}
          value={bundle.data.ai.status || "disabled"}
        />
        <MiniMetric
          icon={<FileWarning className="size-4 text-primary" />}
          label="Warnings"
          note={`${supportFileCount} support file(s)`}
          value={String(warningCount)}
        />
      </div>

      <dl className="space-y-3">
        <PropertyRow label="Bundle" value={bundle.summary.name} />
        <PropertyRow label="Generated" value={formatTimestamp(bundle.summary.generated_at)} />
        <PropertyRow label="Analyzed path" value={bundle.summary.analyzed_path || "local checkout"} />
      </dl>
    </div>
  );
}

function BundleSignalsCard({ bundle }: { bundle: WorkbenchBundle | null }) {
  if (!bundle) {
    return <p className="text-sm text-muted-foreground">Signals appear after a bundle is selected.</p>;
  }

  return (
    <div className="space-y-4">
      <MiniMetric
        icon={<FolderTree className="size-4 text-primary" />}
        label="Modules"
        note="Structured directories"
        value={String(bundle.data.modules.length)}
      />
      <MiniMetric
        icon={<GitBranch className="size-4 text-primary" />}
        label="Reading path"
        note="Suggested entry points"
        value={String(bundle.data.reading_path.length)}
      />
      <div className="space-y-2">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Top languages</p>
        <ul className="space-y-2">
          {bundle.data.languages.slice(0, 4).map((language) => (
            <li className="rounded-2xl border border-border bg-background/60 px-3 py-2 text-sm text-muted-foreground" key={language.name}>
              <span className="font-medium text-foreground">{language.name}</span>
              <span> · {language.file_count} files · {language.line_count} lines</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

function QuickLinksCard({ bundle }: { bundle: WorkbenchBundle | null }) {
  if (!bundle) {
    return <p className="text-sm text-muted-foreground">Quick links will appear when a bundle is selected.</p>;
  }

  const links = [
    { label: "Bundle README", href: bundleLink(bundle.summary.name, "README.md") },
    { label: "Viewer snapshot", href: bundleLink(bundle.summary.name, "ui/index.html") },
    { label: "Architecture Mermaid", href: bundleLink(bundle.summary.name, bundle.data.links.architecture_diagram) },
    { label: "Dependency Mermaid", href: bundleLink(bundle.summary.name, bundle.data.links.dependency_diagram) },
  ];

  return (
    <div className="space-y-2">
      {links.map((link) => (
        <a
          className="flex items-center justify-between rounded-2xl border border-border bg-background/60 px-3 py-3 text-sm text-foreground transition hover:border-border-strong hover:bg-background"
          href={link.href}
          key={link.label}
          rel="noreferrer"
          target="_blank"
        >
          <span>{link.label}</span>
          <Link2 className="size-4 text-muted-foreground" />
        </a>
      ))}
    </div>
  );
}

function MiniMetric({
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
    <article className="rounded-2xl border border-border bg-background/60 p-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{label}</p>
        {icon}
      </div>
      <strong className="mt-3 block text-lg font-semibold text-foreground">{value}</strong>
      <p className="mt-1 text-xs text-muted-foreground">{note}</p>
    </article>
  );
}

function PropertyRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-3">
      <span className="pt-0.5 text-xs text-muted-foreground">{label}</span>
      <span className="max-w-[14rem] text-right text-sm text-foreground">{value}</span>
    </div>
  );
}

function formatTimestamp(value: string) {
  if (!value) {
    return "unknown";
  }

  const timestamp = new Date(value);
  if (Number.isNaN(timestamp.getTime())) {
    return value;
  }

  return timestamp.toLocaleString();
}

function runStatusPillClassName(status: string) {
  switch (status) {
    case "running":
    case "queued":
      return "status-pill status-pill-active";
    case "canceling":
      return "status-pill border-amber-400/35 bg-amber-400/10 text-amber-100";
    case "failed":
    case "canceled":
      return "status-pill border-danger/35 bg-danger/10 text-danger";
    default:
      return "status-pill status-pill-neutral";
  }
}
