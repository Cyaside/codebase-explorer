import type { BundleSummary, WorkbenchStatusResponse, WorkbenchBundle } from "@/lib/types";

interface StatsGridProps {
  bundle: WorkbenchBundle | null;
  bundles: BundleSummary[];
  connections: number;
  status: WorkbenchStatusResponse | null;
}

export function StatsGrid({ bundle, bundles, connections, status }: StatsGridProps) {
  const cards = bundle
    ? [
        { label: "Project", value: bundle.summary.project_name || "Repository", note: bundle.summary.project_type || "Unknown shape" },
        { label: "Files", value: String(bundle.summary.total_files), note: bundle.summary.analyzed_path || "Local checkout" },
        { label: "AI", value: bundle.summary.ai_status || "disabled", note: bundle.data.ai.provider || "No provider" },
        { label: "Changes", value: String(bundle.data.changes.frequently_mentioned_areas.length), note: `${bundle.summary.support_file_count} support file(s)` },
      ]
    : [
        { label: "Projects", value: String(bundles.length), note: "Recent bundles available" },
        { label: "Connections", value: String(connections), note: "Saved locally without API keys" },
        { label: "Output root", value: status?.output_root || "out", note: "Local-first storage" },
        { label: "Mode", value: "Deterministic-first", note: "AI stays optional" },
      ];

  return (
    <section className="grid gap-4 xl:grid-cols-4">
      {cards.map((card) => (
        <article className="rounded-3xl border border-border bg-card/90 p-4 shadow-sm" key={card.label}>
          <p className="text-[10px] uppercase tracking-[0.28em] text-muted-foreground">{card.label}</p>
          <strong className="mt-3 block text-2xl font-bold">{card.value}</strong>
          <p className="mt-2 text-sm text-muted-foreground">{card.note}</p>
        </article>
      ))}
    </section>
  );
}
