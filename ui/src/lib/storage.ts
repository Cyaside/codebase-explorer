import type { ConnectionProfile, PersistedUIState, ProviderProfile, SavedWorkspace } from "@/lib/types";

const PROFILE_STORAGE_KEY = "codearch.workbench.profiles.v3";
const UI_STATE_STORAGE_KEY = "codearch.workbench.state.v2";
const WORKSPACE_STORAGE_KEY = "codearch.workbench.workspaces.v1";

const presetProfiles: ConnectionProfile[] = [
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
  const provider = sanitizeProvider(profile.provider);
  if (!id || !label) {
    return null;
  }
  if (!provider) {
    return null;
  }

  return {
    id,
    label,
    provider,
    locked: Boolean(profile.locked),
  };
}

function sanitizeStringArray(value: unknown) {
  return Array.isArray(value)
    ? value.map((item) => (typeof item === "string" ? item.trim() : "")).filter(Boolean)
    : [];
}

function sanitizeWorkspace(value: unknown): SavedWorkspace | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const workspace = value as Record<string, unknown>;
  const id = typeof workspace.id === "string" ? workspace.id.trim() : "";
  const label = typeof workspace.label === "string" ? workspace.label.trim() : "";
  const repoPath = typeof workspace.repoPath === "string" ? workspace.repoPath.trim() : "";
  if (!id || !label) {
    return null;
  }

  const now = new Date().toISOString();
  return {
    id,
    label,
    repoPath,
    supportFiles: sanitizeStringArray(workspace.supportFiles),
    ignorePatterns: sanitizeStringArray(workspace.ignorePatterns),
    selectedProfile: typeof workspace.selectedProfile === "string" ? workspace.selectedProfile.trim() : presetProfiles[0].id,
    activeBundle: typeof workspace.activeBundle === "string" ? workspace.activeBundle.trim() : "",
    createdAt: typeof workspace.createdAt === "string" ? workspace.createdAt : now,
    updatedAt: typeof workspace.updatedAt === "string" ? workspace.updatedAt : now,
  };
}

function cloneProfile(profile: ConnectionProfile): ConnectionProfile {
  return {
    id: profile.id,
    label: profile.label,
    locked: profile.locked,
    provider: { ...profile.provider },
  };
}

function cloneWorkspace(workspace: SavedWorkspace): SavedWorkspace {
  return {
    ...workspace,
    supportFiles: [...workspace.supportFiles],
    ignorePatterns: [...workspace.ignorePatterns],
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
    provider: {
      name: profile.provider.name,
      model: profile.provider.model,
      baseUrl: profile.provider.baseUrl,
    },
  }));

  window.localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(persisted));
}

export function loadWorkspaces() {
  try {
    const raw = window.localStorage.getItem(WORKSPACE_STORAGE_KEY);
    if (!raw) {
      return [] as SavedWorkspace[];
    }

    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [] as SavedWorkspace[];
    }

    return parsed
      .map(sanitizeWorkspace)
      .filter((workspace): workspace is SavedWorkspace => workspace !== null)
      .map(cloneWorkspace)
      .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt));
  } catch {
    return [] as SavedWorkspace[];
  }
}

export function persistWorkspaces(workspaces: SavedWorkspace[]) {
  const persisted = workspaces.map((workspace) => ({
    ...workspace,
    supportFiles: [...workspace.supportFiles],
    ignorePatterns: [...workspace.ignorePatterns],
  }));

  window.localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify(persisted));
}

export function loadUIState(): PersistedUIState {
  try {
    const raw = window.localStorage.getItem(UI_STATE_STORAGE_KEY);
    if (!raw) {
      return {
        activeTab: "projects",
        activeWorkspace: "",
      };
    }

    const parsed = JSON.parse(raw) as Partial<PersistedUIState>;
    return {
      activeTab: parsed.activeTab ?? "projects",
      activeWorkspace: parsed.activeWorkspace ?? "",
    };
  } catch {
    return {
      activeTab: "projects",
      activeWorkspace: "",
    };
  }
}

export function persistUIState(value: PersistedUIState) {
  window.localStorage.setItem(UI_STATE_STORAGE_KEY, JSON.stringify(value));
}

export function createWorkspace(seed?: Partial<SavedWorkspace>): SavedWorkspace {
  const timestamp = new Date().toISOString();
  return {
    id: `workspace-${Date.now()}`,
    label: seed?.label?.trim() || "Untitled Project",
    repoPath: seed?.repoPath?.trim() || "",
    supportFiles: seed?.supportFiles ? [...seed.supportFiles] : [],
    ignorePatterns: seed?.ignorePatterns ? [...seed.ignorePatterns] : [],
    selectedProfile: seed?.selectedProfile || presetProfiles[0].id,
    activeBundle: seed?.activeBundle || "",
    createdAt: seed?.createdAt || timestamp,
    updatedAt: seed?.updatedAt || timestamp,
  };
}
