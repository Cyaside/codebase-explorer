import type { ConnectionProfile, PersistedUIState, ProviderProfile, SavedWorkspace } from "@/lib/types";

const PROFILE_STORAGE_KEY = "codearch.workbench.connection.v1";
const LEGACY_PROFILE_STORAGE_KEY = "codearch.workbench.profiles.v3";
const UI_STATE_STORAGE_KEY = "codearch.workbench.state.v2";
const WORKSPACE_STORAGE_KEY = "codearch.workbench.workspaces.v1";

const defaultConnection: ConnectionProfile = {
  id: "openai",
  label: "OpenAI-compatible",
  provider: { name: "compatible", model: "", baseUrl: "" },
};

function sanitizeProvider(value: unknown): ProviderProfile | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const provider = value as Record<string, unknown>;
  const storedName = typeof provider.name === "string" ? provider.name.trim().toLowerCase() : "";
  const name = storedName === "openai" || storedName === "openai-compatible" ? "compatible" : storedName;
  const storedBaseUrl = typeof provider.baseUrl === "string" ? provider.baseUrl.trim() : "";
  if (!name) {
    return null;
  }

  return {
    name,
    model: typeof provider.model === "string" ? provider.model.trim() : "",
    baseUrl: storedName === "openai" && !storedBaseUrl ? "https://api.openai.com/v1" : storedBaseUrl,
  };
}

function sanitizeProfile(value: unknown): ConnectionProfile | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const profile = value as Record<string, unknown>;
  const provider = sanitizeProvider(profile.provider);
  if (!provider) {
    return null;
  }

  return {
    id: defaultConnection.id,
    label: defaultConnection.label,
    provider: { ...provider, name: "compatible" },
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
    activeBundle: typeof workspace.activeBundle === "string" ? workspace.activeBundle.trim() : "",
    createdAt: typeof workspace.createdAt === "string" ? workspace.createdAt : now,
    updatedAt: typeof workspace.updatedAt === "string" ? workspace.updatedAt : now,
  };
}

function cloneProfile(profile: ConnectionProfile): ConnectionProfile {
  return {
    id: profile.id,
    label: profile.label,
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

export function defaultProfile() {
  return cloneProfile(defaultConnection);
}

export function loadProfile() {
  try {
    const raw = window.localStorage.getItem(PROFILE_STORAGE_KEY);
    if (raw) {
      return sanitizeProfile(JSON.parse(raw)) || defaultProfile();
    }
    const legacy = JSON.parse(window.localStorage.getItem(LEGACY_PROFILE_STORAGE_KEY) || "[]") as unknown;
    const previous = Array.isArray(legacy) ? legacy.find((item) => item && typeof item === "object" && (item as { id?: string }).id === "openai") : null;
    return sanitizeProfile(previous) || defaultProfile();
  } catch {
    return defaultProfile();
  }
}

export function persistProfile(profile: ConnectionProfile) {
  const persisted = {
    id: defaultConnection.id,
    label: defaultConnection.label,
    provider: {
      name: "compatible",
      model: profile.provider.model,
      baseUrl: profile.provider.baseUrl,
    },
  };

  window.localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(persisted));
  window.localStorage.removeItem(LEGACY_PROFILE_STORAGE_KEY);
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
    activeBundle: seed?.activeBundle || "",
    createdAt: seed?.createdAt || timestamp,
    updatedAt: seed?.updatedAt || timestamp,
  };
}
