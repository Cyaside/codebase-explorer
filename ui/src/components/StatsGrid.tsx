import { Bot, FolderGit2, GitBranch, ShieldAlert } from "lucide-react";

import type { WorkbenchBundle } from "@/lib/types";

interface StatsGridProps {
  bundle: WorkbenchBundle | null;
}

export function StatsGrid({ bundle }: StatsGridProps) {
  const cards = bundle
    ? [
        {
          icon: <FolderGit2 className="size-4" />,
          label: "Core modules",
          note: `${bundle.data.modules.length} mapped modules`,
          value: String(bundle.data.core_modules.length || bundle.data.modules.length),
        },
        {
          icon: <GitBranch className="size-4" />,
          label: "Reading path",
          note: "Recommended checkpoints",
          value: String(bundle.data.reading_path.length),
        },
        {
          icon: <Bot className="size-4" />,
          label: "AI status",
          note: bundle.data.ai.provider || "deterministic",
          value: bundle.data.ai.status || "disabled",
        },
        {
          icon: <ShieldAlert className="size-4" />,
          label: "Issue signals",
          note: `${bundle.data.changes.support_file_count} support file(s) linked`,
          value: String(bundle.data.changes.frequently_mentioned_areas.length),
        },
      ]
    : [
        {
          icon: <FolderGit2 className="size-4" />,
          label: "Workspace",
          note: "Open a project to begin",
          value: "--",
        },
        {
          icon: <GitBranch className="size-4" />,
          label: "Flowchart",
          note: "Graph appears after analysis",
          value: "--",
        },
        {
          icon: <Bot className="size-4" />,
          label: "AI status",
          note: "Optional synthesis layer",
          value: "--",
        },
        {
          icon: <ShieldAlert className="size-4" />,
          label: "Issues",
          note: "Support files shape change awareness",
          value: "--",
        },
      ];

  return (
    <section className="grid gap-3 xl:grid-cols-4">
      {cards.map((card) => (
        <article className="panel-block h-full min-h-[10.5rem]" key={card.label}>
          <div className="flex items-center justify-between gap-3">
            <span className="panel-kicker">{card.label}</span>
            <span className="text-zinc-500">{card.icon}</span>
          </div>
          <strong className="mt-8 block text-4xl font-semibold tracking-tight text-zinc-50">{card.value}</strong>
          <p className="mt-3 text-sm text-zinc-300">{card.note}</p>
        </article>
      ))}
    </section>
  );
}
