import { CopyPlus, KeyRound, Save, Trash2 } from "lucide-react";

import type { ConnectionProfile, SupportedProviderOption } from "@/lib/types";

interface ConnectionPanelProps {
  apiKey: string;
  onAPIKeyChange: (value: string) => void;
  onChange: (profile: ConnectionProfile) => void;
  onDelete: () => void;
  onDuplicate: () => void;
  onSave: () => void;
  profile: ConnectionProfile;
  providerOptions: SupportedProviderOption[];
  validationErrors: string[];
}

export function ConnectionPanel({
  apiKey,
  onAPIKeyChange,
  onChange,
  onDelete,
  onDuplicate,
  onSave,
  profile,
  providerOptions,
  validationErrors,
}: ConnectionPanelProps) {
  const providerName = profile.provider.name;
  const providerMeta = providerOptions.find((item) => item.name === providerName);

  return (
    <section className="rail-section">
      <div>
        <p className="panel-kicker">Connection</p>
        <h3 className="mt-3 text-lg font-semibold text-zinc-100">{profile.label}</h3>
      </div>

      <div className="space-y-3">
        <label className="compact-field">
          <span className="compact-label">Label</span>
          <input
            className="compact-input"
            onChange={(event) => onChange({ ...profile, label: event.target.value })}
            type="text"
            value={profile.label}
          />
        </label>

        <label className="compact-field">
          <span className="compact-label">Provider mode</span>
          <select
            className="compact-input"
            onChange={(event) =>
              onChange({
                ...profile,
                provider: {
                  name: event.target.value,
                  model: profile.provider.model || "",
                  baseUrl: event.target.value === "openai" ? "" : profile.provider.baseUrl || "",
                },
              })
            }
            value={providerName}
          >
            {providerOptions.map((option) => (
              <option key={option.name} value={option.name}>
                {option.name}
              </option>
            ))}
          </select>
        </label>

        <label className="compact-field">
          <span className="compact-label">Model</span>
          <input
            className="compact-input"
            onChange={(event) =>
              onChange({
                ...profile,
                provider: {
                  ...profile.provider,
                  model: event.target.value,
                },
              })
            }
            placeholder="mistral-small-latest"
            type="text"
            value={profile.provider.model}
          />
        </label>

        <label className="compact-field">
          <span className="compact-label">Base URL</span>
          <input
            className="compact-input"
            onChange={(event) =>
              onChange({
                ...profile,
                provider: {
                  ...profile.provider,
                  baseUrl: event.target.value,
                },
              })
            }
            placeholder="https://api.mistral.ai/v1"
            type="text"
            value={profile.provider.baseUrl}
          />
        </label>

        <label className="compact-field">
          <span className="compact-label">API key</span>
          <div className="relative">
            <KeyRound className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-600" />
            <input
              className="compact-input pl-10"
              onChange={(event) => onAPIKeyChange(event.target.value)}
              placeholder="Session memory only"
              type="password"
              value={apiKey}
            />
          </div>
        </label>
      </div>

      {providerMeta ? (
        <div className="rounded-2xl border border-zinc-900 bg-black/50 px-3 py-3 text-xs leading-6 text-zinc-500">
          <p className="font-semibold uppercase tracking-[0.18em] text-zinc-500">Contract</p>
          <p>API key: {providerMeta.requires_api_key ? "required" : "optional"}</p>
          <p>Model: {providerMeta.requires_model ? "required" : "optional"}</p>
          <p>Base URL: {providerMeta.requires_base_url ? "required" : "optional"}</p>
        </div>
      ) : null}

      {validationErrors.length ? (
        <div className="rounded-2xl border border-red-950 bg-red-950/35 px-3 py-3 text-sm leading-6 text-red-200">
          {validationErrors.map((error) => (
            <p key={error}>{error}</p>
          ))}
        </div>
      ) : null}

      <div className="flex flex-wrap gap-2">
        <button className="secondary-control" onClick={onSave} type="button">
          <Save className="size-4" />
          Save
        </button>
        <button className="secondary-control" onClick={onDuplicate} type="button">
          <CopyPlus className="size-4" />
          Duplicate
        </button>
        <button className="secondary-control danger" disabled={profile.locked} onClick={onDelete} type="button">
          <Trash2 className="size-4" />
          Delete
        </button>
      </div>
    </section>
  );
}
