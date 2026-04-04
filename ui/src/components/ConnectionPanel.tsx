import type { ReactNode } from "react";
import { CheckCircle2, CopyPlus, KeyRound, Save, ShieldCheck, Trash2, WandSparkles } from "lucide-react";

import type { ConnectionProfile, SupportedProviderOption } from "@/lib/types";

interface ConnectionPanelProps {
  profile: ConnectionProfile;
  providerOptions: SupportedProviderOption[];
  apiKey: string;
  validationErrors: string[];
  onChange: (profile: ConnectionProfile) => void;
  onAPIKeyChange: (value: string) => void;
  onSave: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
}

export function ConnectionPanel({
  profile,
  providerOptions,
  apiKey,
  validationErrors,
  onChange,
  onAPIKeyChange,
  onSave,
  onDuplicate,
  onDelete,
}: ConnectionPanelProps) {
  const providerName = profile.provider?.name || "";
  const modeLabel = profile.provider ? profile.provider.name : "Deterministic";
  const providerMeta = providerOptions.find((item) => item.name === providerName);
  const apiKeyReady = Boolean(apiKey.trim());
  const modelReady = profile.provider ? !providerMeta?.requires_model || Boolean(profile.provider.model.trim()) : true;
  const baseURLReady = profile.provider ? !providerMeta?.requires_base_url || Boolean(profile.provider.baseUrl.trim()) : true;

  return (
    <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-sm">
      <div className="mb-5 flex items-start justify-between gap-3">
        <div>
          <p className="text-[10px] uppercase tracking-[0.28em] text-primary">Connection</p>
          <h2 className="text-xl font-bold">{profile.label}</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Connection metadata is saved locally, but API keys stay only in memory for the current browser session.
          </p>
        </div>
        <span className={profile.provider ? "status-pill status-pill-active" : "status-pill status-pill-neutral"}>{modeLabel}</span>
      </div>

      <div className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-3">
          <SurfaceFact
            icon={<ShieldCheck className="size-4 text-primary" />}
            label="Secret"
            note="Only kept in memory"
            value={apiKeyReady ? "Loaded" : "Missing"}
          />
          <SurfaceFact
            icon={<WandSparkles className="size-4 text-primary" />}
            label="Model"
            note={providerMeta?.requires_model ? "Required by provider" : "Optional for this mode"}
            value={modelReady ? "Ready" : "Missing"}
          />
          <SurfaceFact
            icon={<CheckCircle2 className="size-4 text-primary" />}
            label="Base URL"
            note={providerMeta?.requires_base_url ? "Required by provider" : "Uses provider default"}
            value={baseURLReady ? "Ready" : "Missing"}
          />
        </div>

        <div className="grid gap-4 xl:grid-cols-2">
          <label className="field">
            <span className="field-label">Label</span>
            <span className="field-description">Use a short operational name. This is what appears in the sidebar and becomes the active connection for the next analyze run.</span>
            <input
              className="field-input"
              onChange={(event) => onChange({ ...profile, label: event.target.value })}
              type="text"
              value={profile.label}
            />
          </label>

          <label className="field">
            <span className="field-label">Provider mode</span>
            <span className="field-description">Keep deterministic mode for the fastest baseline runs. Switch to an AI-backed provider only when you want synthesis on top of the deterministic bundle.</span>
            <select
              className="field-input"
              onChange={(event) =>
                onChange({
                  ...profile,
                  provider: event.target.value
                    ? {
                        name: event.target.value,
                        model: profile.provider?.model || "",
                        baseUrl: profile.provider?.baseUrl || "",
                      }
                    : null,
                })
              }
              value={providerName}
            >
              <option value="">Deterministic only</option>
              {providerOptions.map((option) => (
                <option key={option.name} value={option.name}>
                  {option.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="grid gap-4 xl:grid-cols-2">
          <label className="field">
            <span className="field-label">Model</span>
            <span className="field-description">Use a maintained production model. Leave blank only if the selected provider truly treats model selection as optional.</span>
            <input
              className="field-input"
              onChange={(event) =>
                onChange({
                  ...profile,
                  provider: profile.provider
                    ? {
                        ...profile.provider,
                        model: event.target.value,
                      }
                    : null,
                })
              }
              placeholder="mistral-small-latest"
              type="text"
              value={profile.provider?.model || ""}
            />
          </label>

          <label className="field">
            <span className="field-label">Base URL</span>
            <span className="field-description">For OpenAI-compatible endpoints, use the provider root ending in `/v1`. Deterministic mode ignores this field.</span>
            <input
              className="field-input"
              onChange={(event) =>
                onChange({
                  ...profile,
                  provider: profile.provider
                    ? {
                        ...profile.provider,
                        baseUrl: event.target.value,
                      }
                    : null,
                })
              }
              placeholder="https://api.openai.com/v1"
              type="text"
              value={profile.provider?.baseUrl || ""}
            />
          </label>
        </div>

        <label className="field">
          <span className="field-label">API key</span>
          <span className="field-description">Secrets stay in memory for this browser session and are never written into analysis bundles.</span>
          <div className="relative">
            <KeyRound className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <input
              className="field-input pl-10"
              onChange={(event) => onAPIKeyChange(event.target.value)}
              placeholder="Kept only in memory until this browser tab closes"
              type="password"
              value={apiKey}
            />
          </div>
        </label>

        {profile.provider && (
          <div className="rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm text-muted-foreground">
            <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Provider contract</p>
            {providerMeta ? (
              <ul className="space-y-1">
                <li>Requires API key: {providerMeta.requires_api_key ? "yes" : "no"}</li>
                <li>Requires model: {providerMeta.requires_model ? "yes" : "no"}</li>
                <li>Requires base URL: {providerMeta.requires_base_url ? "yes" : "no"}</li>
              </ul>
            ) : (
              <p>This provider is not currently advertised by the backend.</p>
            )}
          </div>
        )}

        {validationErrors.length ? (
          <div className="rounded-2xl border border-danger/40 bg-danger/10 px-4 py-3 text-sm text-danger-foreground">
            <p className="mb-2 font-semibold text-danger">Connection needs attention</p>
            <ul className="space-y-1 text-danger/90">
              {validationErrors.map((error) => (
                <li key={error}>- {error}</li>
              ))}
            </ul>
          </div>
        ) : null}

        <div className="flex flex-wrap gap-3 border-t border-border pt-4">
          <button className="action-button" onClick={onSave} type="button">
            <Save className="size-4" />
            Save Connection
          </button>
          <button className="secondary-button" onClick={onDuplicate} type="button">
            <CopyPlus className="size-4" />
            Duplicate
          </button>
          <button className="secondary-button danger" disabled={profile.locked} onClick={onDelete} type="button">
            <Trash2 className="size-4" />
            Delete
          </button>
        </div>
      </div>
    </section>
  );
}

function SurfaceFact({
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
    <article className="surface-fact">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{label}</span>
        {icon}
      </div>
      <strong className="mt-3 block text-base font-semibold text-foreground">{value}</strong>
      <p className="mt-1 text-xs text-muted-foreground">{note}</p>
    </article>
  );
}
