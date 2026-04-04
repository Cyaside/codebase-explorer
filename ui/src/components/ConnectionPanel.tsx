import { CopyPlus, KeyRound, Save, Trash2 } from "lucide-react";

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
        <div className="grid gap-4 xl:grid-cols-2">
          <label className="field">
            <span className="field-label">Label</span>
            <input
              className="field-input"
              onChange={(event) => onChange({ ...profile, label: event.target.value })}
              type="text"
              value={profile.label}
            />
          </label>

          <label className="field">
            <span className="field-label">Provider mode</span>
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
