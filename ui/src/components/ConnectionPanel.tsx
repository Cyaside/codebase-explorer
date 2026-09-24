import { Activity, KeyRound, ListChecks, Save } from "lucide-react";

import type { ConnectionProfile, ProviderDiagnosticsState } from "@/lib/types";

interface ConnectionPanelProps {
  apiKey: string;
  hasSavedKey: boolean;
  diagnostics: ProviderDiagnosticsState;
  onAPIKeyChange: (value: string) => void;
  onChange: (profile: ConnectionProfile) => void;
  onClearKey: () => void;
  onListModels: () => void;
  onSave: () => void;
  onTestProvider: () => void;
  profile: ConnectionProfile;
  validationErrors: string[];
}

export function ConnectionPanel({
  apiKey,
  hasSavedKey,
  diagnostics,
  onAPIKeyChange,
  onChange,
  onClearKey,
  onListModels,
  onSave,
  onTestProvider,
  profile,
  validationErrors,
}: ConnectionPanelProps) {
  return (
    <section className="rail-section connection-panel">
      <div>
        <p className="panel-kicker">Connection</p>
        <h3 className="mt-3 text-lg font-semibold text-zinc-100">OpenAI-compatible</h3>
        <p className="mt-2 text-sm text-zinc-500">Use any endpoint that supports the OpenAI Chat Completions API.</p>
      </div>

      <div className="space-y-3">
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
            placeholder="Your model ID"
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
            placeholder="https://your-provider.example/v1"
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
              placeholder={hasSavedKey ? "Saved in local backend; enter to replace" : "Enter API key"}
              type="password"
              value={apiKey}
            />
          </div>
        </label>
        <p className="text-xs text-zinc-500">{hasSavedKey ? "Key saved on this computer. It is never sent back to the browser." : "Save to store the key in the local backend, outside this repository."}</p>
      </div>

      {validationErrors.length ? (
        <div className="rounded-2xl border border-red-950 bg-red-950/35 px-3 py-3 text-sm leading-6 text-red-200">
          {validationErrors.map((error) => (
            <p key={error}>{error}</p>
          ))}
        </div>
      ) : null}

      <div className="rounded-2xl border border-zinc-900 bg-black/50 px-3 py-3">
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className="font-semibold uppercase tracking-[0.18em] text-zinc-500 text-xs">Diagnostics</p>
            <p className="mt-2 text-xs leading-5 text-zinc-500">Test the selected endpoint, key, and model before running analysis.</p>
          </div>
          {diagnostics.busy ? <span className="run-pill run-pill-busy">{diagnostics.mode}</span> : null}
        </div>

        <div className="mt-3 flex flex-wrap gap-2">
          <button className="secondary-control" disabled={diagnostics.busy} onClick={onListModels} type="button">
            <ListChecks className="size-4" />
            List models
          </button>
          <button className="secondary-control" disabled={diagnostics.busy} onClick={onTestProvider} type="button">
            <Activity className="size-4" />
            Test model
          </button>
        </div>

        {diagnostics.error ? <p className="mt-3 rounded-2xl border border-red-950 bg-red-950/35 px-3 py-2 text-sm text-red-200">{diagnostics.error}</p> : null}
        {diagnostics.test ? (
          <div className="mt-3 rounded-2xl border border-zinc-900 bg-black px-3 py-3 text-xs leading-6 text-zinc-400">
            <p className="font-semibold text-zinc-200">{diagnostics.test.status} - {diagnostics.test.model}</p>
            <p>Latency: {diagnostics.test.latency_ms}ms</p>
            <p className="mt-2 text-zinc-300">{diagnostics.test.output}</p>
          </div>
        ) : null}
        {diagnostics.models ? (
          <div className="mt-3 rounded-2xl border border-zinc-900 bg-black px-3 py-3 text-xs leading-6 text-zinc-400">
            <p className="font-semibold text-zinc-200">{diagnostics.models.count} model(s) returned in {diagnostics.models.latency_ms}ms</p>
            <div className="mt-2 max-h-40 overflow-y-auto pr-1">
              {diagnostics.models.models.slice(0, 80).map((model) => (
                <button
                  className={`block w-full truncate rounded-xl px-2 py-1 text-left hover:bg-zinc-950 ${
                    model === profile.provider.model ? "text-zinc-50" : "text-zinc-500"
                  }`}
                  key={model}
                  onClick={() =>
                    onChange({
                      ...profile,
                      provider: {
                        ...profile.provider,
                        model,
                      },
                    })
                  }
                  type="button"
                >
                  {model}
                </button>
              ))}
            </div>
          </div>
        ) : null}
      </div>

      <div className="flex flex-wrap gap-2">
        <button className="primary-control" onClick={onSave} type="button">
          <Save className="size-4" />
          Save
        </button>
        {hasSavedKey ? <button className="secondary-control danger" onClick={onClearKey} type="button">Clear saved key</button> : null}
      </div>
    </section>
  );
}
