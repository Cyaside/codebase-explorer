import type { ReactNode } from "react";
import { Database, FolderKanban, Layers3, PlugZap, RefreshCw } from "lucide-react";

import type { BundleSummary, ConnectionProfile, WorkbenchStatusResponse } from "@/lib/types";
import { cn } from "@/lib/utils";

interface SidebarProps {
  profiles: ConnectionProfile[];
  selectedProfile: string;
  onSelectProfile: (profileID: string) => void;
  bundles: BundleSummary[];
  selectedBundle: string;
  onSelectBundle: (bundleName: string) => void;
  onRefresh: () => void;
  status: WorkbenchStatusResponse | null;
}

export function Sidebar({
  profiles,
  selectedProfile,
  onSelectProfile,
  bundles,
  selectedBundle,
  onSelectBundle,
  onRefresh,
  status,
}: SidebarProps) {
  return (
    <aside className="border-r border-border/80 bg-sidebar px-4 py-5">
      <div className="mb-6 flex items-center gap-3">
        <div className="grid size-12 place-items-center rounded-2xl border border-border bg-card text-lg font-bold text-primary shadow-sm">
          C
        </div>
        <div>
          <p className="text-[10px] uppercase tracking-[0.32em] text-primary">Codebase Explorer</p>
          <h1 className="text-lg font-bold text-foreground">Workbench</h1>
        </div>
      </div>

      <div className="space-y-6">
        <SidebarSection
          icon={<PlugZap className="size-4 text-muted-foreground" />}
          title="Connections"
          action={
            <button className="focus-shell text-xs text-muted-foreground transition hover:text-foreground" onClick={onRefresh} type="button">
              sync
            </button>
          }
        >
          <div className="space-y-2">
            {profiles.map((profile) => {
              const selected = profile.id === selectedProfile;
              const meta = profile.provider
                ? `${profile.provider.name}${profile.provider.model ? ` · ${profile.provider.model}` : ""}`
                : "Deterministic only";

              return (
                <button
                  className={cn(
                    "focus-shell w-full rounded-xl border px-3 py-3 text-left transition",
                    selected
                      ? "border-primary/40 bg-accent text-accent-foreground"
                      : "border-border bg-card/80 text-foreground hover:border-border-strong hover:bg-card",
                  )}
                  key={profile.id}
                  onClick={() => onSelectProfile(profile.id)}
                  type="button"
                >
                  <strong className="block text-sm font-semibold">{profile.label}</strong>
                  <span className="mt-1 block text-xs text-muted-foreground">{meta}</span>
                </button>
              );
            })}
          </div>
        </SidebarSection>

        <SidebarSection
          icon={<FolderKanban className="size-4 text-muted-foreground" />}
          title="Recent Bundles"
          action={
            <button className="focus-shell text-xs text-muted-foreground transition hover:text-foreground" onClick={onRefresh} type="button">
              <RefreshCw className="size-3.5" />
            </button>
          }
        >
          <div className="space-y-2">
            {bundles.length ? (
              bundles.map((bundle) => (
                <button
                  className={cn(
                    "focus-shell w-full rounded-xl border px-3 py-3 text-left transition",
                    bundle.name === selectedBundle
                      ? "border-primary/40 bg-accent text-accent-foreground"
                      : "border-border bg-card/80 text-foreground hover:border-border-strong hover:bg-card",
                  )}
                  key={bundle.name}
                  onClick={() => onSelectBundle(bundle.name)}
                  type="button"
                >
                  <strong className="block text-sm font-semibold">{bundle.project_name || bundle.name}</strong>
                  <span className="mt-1 block text-xs text-muted-foreground">
                    {bundle.ai_status || "deterministic"} · {bundle.total_files} files
                  </span>
                </button>
              ))
            ) : (
              <p className="rounded-xl border border-dashed border-border px-3 py-4 text-sm text-muted-foreground">
                No bundle yet. Run your first analysis.
              </p>
            )}
          </div>
        </SidebarSection>

        <SidebarSection icon={<Database className="size-4 text-muted-foreground" />} title="System">
          <div className="space-y-2 rounded-2xl border border-border bg-card/70 p-3">
            <FactRow label="Version" value={status?.app_version || "dev"} />
            <FactRow label="Output root" value={status?.output_root || "out"} />
            <FactRow label="Cache" value={status?.cache_enabled ? status.cache_root : "disabled"} />
            <FactRow
              label="Providers"
              value={status?.supported_providers.length ? status.supported_providers.map((item) => item.name).join(", ") : "none"}
            />
          </div>
        </SidebarSection>

        {status?.bundle_warnings?.length ? (
          <SidebarSection icon={<Database className="size-4 text-danger" />} title="Bundle warnings">
            <div className="space-y-2 rounded-2xl border border-danger/35 bg-danger/10 p-3">
              {status.bundle_warnings.map((warning) => (
                <p className="text-sm text-danger/90" key={warning}>
                  {warning}
                </p>
              ))}
            </div>
          </SidebarSection>
        ) : null}

        <SidebarSection icon={<Layers3 className="size-4 text-muted-foreground" />} title="Mode">
          <p className="rounded-2xl border border-border bg-card/70 p-3 text-sm text-muted-foreground">
            Local-first workbench. Deterministic analysis stays usable even when AI providers are unavailable.
          </p>
        </SidebarSection>
      </div>
    </aside>
  );
}

function SidebarSection({
  action,
  children,
  icon,
  title,
}: {
  action?: ReactNode;
  children: ReactNode;
  icon: ReactNode;
  title: string;
}) {
  return (
    <section>
      <div className="mb-3 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          {icon}
          <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">{title}</p>
        </div>
        {action}
      </div>
      {children}
    </section>
  );
}

function FactRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-xs">
      <span className="text-muted-foreground">{label}</span>
      <strong className="max-w-[11rem] truncate text-right text-foreground">{value}</strong>
    </div>
  );
}
