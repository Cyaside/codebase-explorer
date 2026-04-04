import type { ConnectionProfile, PersistedUIState, ProviderProfile } from "@/lib/types";

const PROFILE_STORAGE_KEY = "codearch.workbench.profiles.v2";
const UI_STATE_STORAGE_KEY = "codearch.workbench.state.v1";

const presetProfiles: ConnectionProfile[] = [
  { id: "deterministic", label: "Deterministic only", provider: null, locked: true },
  { id: "openai", label: "OpenAI", provider: { name: "openai", model: "gpt-4.1-mini", baseUrl: "" } },
  {
    id: "openrouter",
    label: "OpenRouter",
    provider: { name: "openai-compatible", model: "openai/gpt-4.1-mini", baseUrl: "https://openrouter.ai/api/v1" },
  },
  {
    id: "mistral",
    label: "Mistral",
    provider: { name: "openai-compatible", model: "mistral-small-latest", baseUrl: "https://api.mistral.ai/v1" },
  },
];

function sanitizeProvider(value: unknown): ProviderProfile | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const provider = value as Record<string, unknown>;
  const name = typeof provider.name === "string" ? provider.name.trim() : "";
  if (!name) {
    return null;
  }

  return {
    name,
    model: typeof provider.model === "string" ? provider.model.trim() : "",
    baseUrl: typeof provider.baseUrl === "string" ? provider.baseUrl.trim() : "",
  };
}

function sanitizeProfile(value: unknown): ConnectionProfile | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const profile = value as Record<string, unknown>;
  const id = typeof profile.id === "string" ? profile.id.trim() : "";
  const label = typeof profile.label === "string" ? profile.label.trim() : "";
  if (!id || !label) {
    return null;
  }

  return {
    id,
    label,
    provider: sanitizeProvider(profile.provider),
    locked: Boolean(profile.locked),
  };
}

function cloneProfile(profile: ConnectionProfile): ConnectionProfile {
  return {
    id: profile.id,
    label: profile.label,
    locked: profile.locked,
    provider: profile.provider ? { ...profile.provider } : null,
  };
}

export function defaultProfiles() {
  return presetProfiles.map(cloneProfile);
}

export function loadProfiles() {
  try {
    const raw = window.localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) {
      return defaultProfiles();
    }

    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return defaultProfiles();
    }

    const sanitized = parsed
      .map(sanitizeProfile)
      .filter((profile): profile is ConnectionProfile => profile !== null);
    const merged = sanitized.map(cloneProfile);
    const seen = new Set(merged.map((profile) => profile.id));

    for (const profile of presetProfiles) {
      if (!seen.has(profile.id)) {
        merged.unshift(cloneProfile(profile));
      }
    }

    return merged.length ? merged : defaultProfiles();
  } catch {
    return defaultProfiles();
  }
}

export function persistProfiles(profiles: ConnectionProfile[]) {
  const persisted = profiles.map((profile) => ({
    id: profile.id,
    label: profile.label,
    locked: profile.locked,
    provider: profile.provider
      ? {
          name: profile.provider.name,
          model: profile.provider.model,
          baseUrl: profile.provider.baseUrl,
        }
      : null,
  }));

  window.localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(persisted));
}

export function loadUIState(): PersistedUIState {
  try {
    const raw = window.localStorage.getItem(UI_STATE_STORAGE_KEY);
    if (!raw) {
      return {
        activeTab: "summary",
        selectedBundle: "",
        selectedProfile: "",
      };
    }

    const parsed = JSON.parse(raw) as Partial<PersistedUIState>;
    return {
      activeTab: parsed.activeTab ?? "summary",
      selectedBundle: parsed.selectedBundle ?? "",
      selectedProfile: parsed.selectedProfile ?? "",
    };
  } catch {
    return {
      activeTab: "summary",
      selectedBundle: "",
      selectedProfile: "",
    };
  }
}

export function persistUIState(value: PersistedUIState) {
  window.localStorage.setItem(UI_STATE_STORAGE_KEY, JSON.stringify(value));
}
