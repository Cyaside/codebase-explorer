import { startTransition, useEffect, useRef, useState } from "react";
import { Command, FolderOpenDot, RefreshCw } from "lucide-react";

import { CommandPalette, type CommandPaletteAction } from "@/components/CommandPalette";
import { Sidebar } from "@/components/Sidebar";
import { TabPanels } from "@/components/TabPanels";
import {
  buildProviderPayload,
  cancelAnalyzeRun,
  deleteBundle,
  fetchAnalyzeRun,
  fetchBundle,
  fetchProviderModels,
  fetchStatus,
  mergeBundleSummary,
  startAnalyzeRun,
  testProvider,
} from "@/lib/api";
import { useWorkbenchShortcuts } from "@/hooks/useWorkbenchShortcuts";
import { createWorkspace, defaultProfiles, loadProfiles, loadUIState, loadWorkspaces, persistProfiles, persistUIState, persistWorkspaces } from "@/lib/storage";
import { openAnalyzeRunStream } from "@/lib/stream";
import type {
  AnalyzeRun,
  ConnectionProfile,
  InspectorState,
  ProviderDiagnosticsState,
  SavedWorkspace,
  TabKey,
  WorkbenchBundle,
  WorkbenchStatusResponse,
} from "@/lib/types";
import { bundleLink, normalizeLocalPath } from "@/lib/utils";
import { validateProfile } from "@/lib/validation";

const initialUIState = loadUIState();
const initialProfiles = loadProfiles();
const initialWorkspaces = loadWorkspaces();
const workbenchTabs: TabKey[] = [
  "projects",
  "project",
  "connections",
  "properties",
  "dashboard",
  "summary",
  "architecture",
  "flowchart",
  "issues",
  "recommendations",
];

const initialProviderDiagnostics: ProviderDiagnosticsState = {
  busy: false,
  mode: "",
  error: "",
  models: null,
  test: null,
};

export function App() {
  const [status, setStatus] = useState<WorkbenchStatusResponse | null>(null);
  const [bundles, setBundles] = useState<WorkbenchStatusResponse["recent_bundles"]>([]);
  const [bundleCache, setBundleCache] = useState<Record<string, WorkbenchBundle>>({});
  const [profiles, setProfiles] = useState(initialProfiles);
  const [workspaces, setWorkspaces] = useState(initialWorkspaces);
  const [activeWorkspaceID, setActiveWorkspaceID] = useState(initialUIState.activeWorkspace || initialWorkspaces[0]?.id || "");
  const [activeTab, setActiveTab] = useState<TabKey>(initialUIState.activeTab || "projects");
  const [profileSecrets, setProfileSecrets] = useState<Record<string, string>>({});
  const [activeRun, setActiveRun] = useState<AnalyzeRun | null>(null);
  const [busyDetail, setBusyDetail] = useState("");
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [toastMessage, setToastMessage] = useState("");
  const [inspector, setInspector] = useState<InspectorState | null>(null);
  const [providerDiagnostics, setProviderDiagnostics] = useState<ProviderDiagnosticsState>(initialProviderDiagnostics);

  const repoInputRef = useRef<HTMLInputElement | null>(null);

  const activeWorkspace = workspaces.find((workspace) => workspace.id === activeWorkspaceID) || null;
  const activeProfileID = activeWorkspace?.selectedProfile || profiles[0]?.id || "";
  const profile = profiles.find((item) => item.id === activeProfileID) || profiles[0] || defaultProfiles()[0];
  const providerOptions = status?.supported_providers || [];
  const apiKey = profile ? profileSecrets[profile.id] || "" : "";
  const validationErrors = profile ? validateProfile(profile, apiKey, providerOptions) : [];
  const workspaceBundles = activeWorkspace ? bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, activeWorkspace.repoPath)) : [];
  const selectedBundleName = activeWorkspace?.activeBundle || workspaceBundles[0]?.name || "";
  const currentBundle = selectedBundleName ? bundleCache[selectedBundleName] || null : null;
  const busy = !!activeRun && ["queued", "running", "canceling"].includes(activeRun.status);

  useEffect(() => {
    void refreshStatus();
  }, []);

  useEffect(() => {
    persistProfiles(profiles);
  }, [profiles]);

  useEffect(() => {
    persistWorkspaces(workspaces);
  }, [workspaces]);

  useEffect(() => {
    persistUIState({
      activeTab,
      activeWorkspace: activeWorkspaceID,
    });
  }, [activeTab, activeWorkspaceID]);

  useEffect(() => {
    if (!activeWorkspace && workspaces.length) {
      setActiveWorkspaceID(workspaces[0].id);
    }
  }, [activeWorkspace, workspaces]);

  useEffect(() => {
    if (!activeWorkspace || profiles.some((item) => item.id === activeWorkspace.selectedProfile)) {
      return;
    }

    updateWorkspace(activeWorkspace.id, {
      selectedProfile: profiles[0]?.id || "",
      updatedAt: new Date().toISOString(),
    });
  }, [activeWorkspace?.id, activeWorkspace?.selectedProfile, profiles]);

  useEffect(() => {
    if (!toastMessage) {
      return;
    }

    const timeoutID = window.setTimeout(() => setToastMessage(""), 3200);
    return () => window.clearTimeout(timeoutID);
  }, [toastMessage]);

  useEffect(() => {
    if (!activeWorkspace) {
      return;
    }

    const nextBundleName = activeWorkspace.activeBundle || workspaceBundles[0]?.name;
    if (nextBundleName && !bundleCache[nextBundleName]) {
      void loadBundle(nextBundleName, activeWorkspace.id);
    }
  }, [activeWorkspace?.activeBundle, activeWorkspace?.id, workspaceBundles.length]);

  useEffect(() => {
    if (!activeRun || !["queued", "running", "canceling"].includes(activeRun.status)) {
      return;
    }

    let canceled = false;
    let intervalID = 0;

    const pollOnce = async () => {
      try {
        const response = await fetchAnalyzeRun(activeRun.id);
        if (!canceled) {
          applyRunSnapshot(response.run);
        }
      } catch (error) {
        if (!canceled) {
          setErrorMessage(error instanceof Error ? error.message : "Failed to poll analyze run.");
        }
      }
    };

    const startPolling = () => {
      if (intervalID) {
        return;
      }

      intervalID = window.setInterval(() => {
        void pollOnce();
      }, 700);
      void pollOnce();
    };

    const stopStream = openAnalyzeRunStream(activeRun.id, {
      onRun: (run) => {
        if (!canceled) {
          applyRunSnapshot(run);
        }
      },
      onFallback: () => {
        if (!canceled) {
          startPolling();
        }
      },
    });

    return () => {
      canceled = true;
      stopStream();
      if (intervalID) {
        window.clearInterval(intervalID);
      }
    };
  }, [activeRun?.id, activeRun?.status]);

  useEffect(() => {
    setInspector(buildDefaultInspector(activeWorkspace, currentBundle, activeTab));
  }, [activeWorkspace?.id, currentBundle?.summary.name, activeTab]);

  useWorkbenchShortcuts({
    activeTab,
    busy,
    commandPaletteOpen,
    onNextBundle: () => {
      void selectRelativeBundle(1);
    },
    onPreviousBundle: () => {
      void selectRelativeBundle(-1);
    },
    onAnalyze: () => {
      void handleAnalyze();
    },
    onCancel: () => {
      void handleCancelAnalyze();
    },
    onCloseCommandPalette: () => {
      setCommandPaletteOpen(false);
    },
    onFocusAnalyze: () => {
      repoInputRef.current?.focus();
      repoInputRef.current?.select();
    },
    onOpenCommandPalette: () => {
      setCommandPaletteOpen(true);
    },
    onRefresh: () => {
      void refreshStatus();
    },
    onSelectTab: setActiveTab,
    tabs: workbenchTabs,
  });

  const commandActions = buildCommandActions({
    bundles: workspaceBundles.length ? workspaceBundles : bundles.slice(0, 8),
    onAnalyze: () => {
      void handleAnalyze();
    },
    onCancelAnalyze: () => {
      void handleCancelAnalyze();
    },
    onCreateWorkspace: handleCreateWorkspace,
    onFocusAnalyze: () => {
      repoInputRef.current?.focus();
      repoInputRef.current?.select();
    },
    onRefresh: () => {
      void refreshStatus();
    },
    onSelectBundle: (bundleName) => {
      void loadBundle(bundleName);
    },
    onSelectTab: setActiveTab,
    onSelectWorkspace: (workspaceID) => {
      void selectWorkspace(workspaceID);
    },
    tabs: workbenchTabs,
    workspaces,
    busy,
  });

  async function refreshStatus(preferredBundleName?: string) {
    try {
      setErrorMessage("");
      const nextStatus = await fetchStatus();

      startTransition(() => {
        setStatus(nextStatus);
        setBundles(nextStatus.recent_bundles || []);
      });

      const scopedBundles = activeWorkspace
        ? nextStatus.recent_bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, activeWorkspace.repoPath))
        : nextStatus.recent_bundles;
      const nextBundleName =
        preferredBundleName ||
        (activeWorkspace?.activeBundle && scopedBundles.some((bundle) => bundle.name === activeWorkspace.activeBundle)
          ? activeWorkspace.activeBundle
          : scopedBundles[0]?.name || "");

      if (nextBundleName) {
        await loadBundle(nextBundleName, activeWorkspace?.id);
      }
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to load workbench status.");
    }
  }

  async function loadBundle(bundleName: string, workspaceID = activeWorkspace?.id || activeWorkspaceID) {
    try {
      setErrorMessage("");
      if (!bundleCache[bundleName]) {
        const bundle = await fetchBundle(bundleName);
        setBundleCache((current) => ({
          ...current,
          [bundleName]: bundle,
        }));
      }

      if (workspaceID) {
        updateWorkspace(workspaceID, {
          activeBundle: bundleName,
          updatedAt: new Date().toISOString(),
        });
      }
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to load bundle.");
    }
  }

  async function selectRelativeBundle(direction: -1 | 1) {
    if (!workspaceBundles.length) {
      return;
    }

    const currentIndex = workspaceBundles.findIndex((bundle) => bundle.name === selectedBundleName);
    const startIndex = currentIndex >= 0 ? currentIndex : 0;
    const nextIndex = (startIndex + direction + workspaceBundles.length) % workspaceBundles.length;
    await loadBundle(workspaceBundles[nextIndex].name, activeWorkspace?.id);
  }

  async function selectWorkspace(workspaceID: string, sourceWorkspaces = workspaces, sourceBundles = bundles) {
    setActiveWorkspaceID(workspaceID);

    const workspace = sourceWorkspaces.find((item) => item.id === workspaceID);
    if (!workspace) {
      return;
    }

    const nextWorkspaceBundles = sourceBundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, workspace.repoPath));
    const nextBundleName =
      workspace.activeBundle && nextWorkspaceBundles.some((bundle) => bundle.name === workspace.activeBundle)
        ? workspace.activeBundle
        : nextWorkspaceBundles[0]?.name || "";

    if (nextBundleName) {
      await loadBundle(nextBundleName, workspaceID);
    } else if (workspace.activeBundle) {
      updateWorkspace(workspaceID, {
        activeBundle: "",
        updatedAt: new Date().toISOString(),
      });
    }
  }

  function handleCreateWorkspace() {
    const nextWorkspace = createWorkspace({
      selectedProfile: profiles[0]?.id || "",
    });

    setWorkspaces((current) => [nextWorkspace, ...current]);
    setActiveWorkspaceID(nextWorkspace.id);
    setActiveTab("project");
    setToastMessage(`Created workspace "${nextWorkspace.label}".`);
    window.setTimeout(() => {
      repoInputRef.current?.focus();
      repoInputRef.current?.select();
    }, 10);
  }

  function updateWorkspace(workspaceID: string, patch: Partial<SavedWorkspace>) {
    setWorkspaces((current) =>
      current.map((workspace) => (workspace.id === workspaceID ? { ...workspace, ...patch } : workspace)).sort(sortWorkspacesByUpdatedAt),
    );
  }

  function applyRunSnapshot(run: AnalyzeRun) {
    setActiveRun(run);
    setBusyDetail(run.progress[run.progress.length - 1]?.detail || run.status);

    if (run.status === "succeeded" && run.response) {
      const response = run.response;
      setErrorMessage("");
      setBundleCache((current) => ({
        ...current,
        [response.bundle.name]: {
          summary: response.bundle,
          data: response.data,
        },
      }));
      setBundles((current) => mergeBundleSummary(current, response.bundle));

      if (activeWorkspace) {
        updateWorkspace(activeWorkspace.id, {
          activeBundle: response.bundle.name,
          label: activeWorkspace.label === "Untitled Project" ? response.bundle.project_name || activeWorkspace.label : activeWorkspace.label,
          updatedAt: new Date().toISOString(),
        });
      }

      setToastMessage(`Analysis ready for ${response.bundle.project_name || response.result.project_name || "repository"}.`);
      void refreshStatus(response.bundle.name);
      return;
    }

    if (run.status === "failed") {
      setErrorMessage(run.error || "Analyze run failed.");
      return;
    }

    if (run.status === "canceled") {
      setToastMessage("Analyze run canceled.");
    }
  }

  function handleSaveProfile() {
    if (validationErrors.length) {
      setErrorMessage(validationErrors.join(" "));
      return;
    }

    setErrorMessage("");
    setToastMessage(`Saved connection "${profile.label}". API keys remain only in session memory.`);
  }

  async function handleListProviderModels() {
    const blockingErrors = validationErrors.filter((error) => !error.toLowerCase().includes("model is required"));
    if (blockingErrors.length) {
      setErrorMessage(blockingErrors.join(" "));
      return;
    }

    setErrorMessage("");
    setProviderDiagnostics((current) => ({
      ...current,
      busy: true,
      error: "",
      mode: "models",
      models: null,
    }));
    try {
      const response = await fetchProviderModels({
        provider: buildProviderPayload(profile, apiKey),
      });
      setProviderDiagnostics((current) => ({
        ...current,
        busy: false,
        error: "",
        mode: "",
        models: response,
      }));
      setToastMessage(`Loaded ${response.count} model(s) from ${profile.label}.`);
    } catch (error) {
      setProviderDiagnostics((current) => ({
        ...current,
        busy: false,
        error: error instanceof Error ? error.message : "Failed to list provider models.",
        mode: "",
        models: null,
      }));
    }
  }

  async function handleTestProvider() {
    if (validationErrors.length) {
      setErrorMessage(validationErrors.join(" "));
      return;
    }

    setErrorMessage("");
    setProviderDiagnostics((current) => ({
      ...current,
      busy: true,
      error: "",
      mode: "test",
      test: null,
    }));
    try {
      const response = await testProvider({
        provider: buildProviderPayload(profile, apiKey),
      });
      setProviderDiagnostics((current) => ({
        ...current,
        busy: false,
        error: "",
        mode: "",
        test: response,
      }));
      setToastMessage(`Provider test succeeded with ${response.model}.`);
    } catch (error) {
      setProviderDiagnostics((current) => ({
        ...current,
        busy: false,
        error: error instanceof Error ? error.message : "Provider test failed.",
        mode: "",
        test: null,
      }));
    }
  }

  function handleDuplicateProfile() {
    const cloneID = `profile-${Date.now()}`;
    const clone: ConnectionProfile = {
      id: cloneID,
      label: `${profile.label} Copy`,
      provider: { ...profile.provider },
    };

    setProfiles((current) => [...current, clone]);
    if (activeWorkspace) {
      updateWorkspace(activeWorkspace.id, {
        selectedProfile: cloneID,
        updatedAt: new Date().toISOString(),
      });
    }
    setProfileSecrets((current) => ({
      ...current,
      [cloneID]: current[profile.id] || "",
    }));
    setToastMessage(`Duplicated "${profile.label}".`);
  }

  function handleDeleteProfile() {
    if (profiles.length <= 1) {
      setErrorMessage("Keep at least one connection profile available.");
      return;
    }
    if (profile.locked) {
      setErrorMessage("Preset profiles cannot be deleted.");
      return;
    }

    const fallbackProfileID = profiles.find((item) => item.id !== profile.id)?.id || profiles[0]?.id || "";
    setProfiles((current) => current.filter((item) => item.id !== profile.id));
    setProfileSecrets((current) => {
      const next = { ...current };
      delete next[profile.id];
      return next;
    });
    setWorkspaces((current) =>
      current.map((workspace) =>
        workspace.selectedProfile === profile.id ? { ...workspace, selectedProfile: fallbackProfileID, updatedAt: new Date().toISOString() } : workspace,
      ),
    );
    setToastMessage(`Deleted "${profile.label}".`);
  }

  async function handleAnalyze() {
    if (!activeWorkspace) {
      setErrorMessage("Create or select a workspace first.");
      return;
    }

    const repoPath = activeWorkspace.repoPath.trim();
    if (!repoPath) {
      setErrorMessage("Repository path is required.");
      return;
    }
    if (validationErrors.length) {
      setErrorMessage(validationErrors.join(" "));
      return;
    }

    setBusyDetail("Submitting analyze request...");
    setErrorMessage("");

    try {
      const response = await startAnalyzeRun({
        repo_path: repoPath,
        deterministic_only: false,
        ai_mode: "full-ai",
        ai_read_budget: 24,
        ai_token_budget: 32000,
        support_files: activeWorkspace.supportFiles,
        extra_ignore_patterns: activeWorkspace.ignorePatterns,
        provider: buildProviderPayload(profile, apiKey),
      });
      applyRunSnapshot(response.run);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Analyze request failed.");
    }
  }

  async function handleCancelAnalyze() {
    if (!activeRun || !busy) {
      return;
    }

    try {
      const response = await cancelAnalyzeRun(activeRun.id);
      applyRunSnapshot(response.run);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to cancel analyze run.");
    }
  }

  async function handleDeleteBundle(bundleName: string) {
    const nextBundles = bundles.filter((bundle) => bundle.name !== bundleName);

    try {
      setErrorMessage("");
      await deleteBundle(bundleName);

      setBundleCache((current) => {
        const next = { ...current };
        delete next[bundleName];
        return next;
      });
      setBundles(nextBundles);
      setWorkspaces((current) =>
        current.map((workspace) => {
          if (workspace.activeBundle !== bundleName) {
            return workspace;
          }

          const workspaceScopedBundles = nextBundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, workspace.repoPath));
          return {
            ...workspace,
            activeBundle: workspaceScopedBundles[0]?.name || "",
            updatedAt: new Date().toISOString(),
          };
        }),
      );
      setToastMessage(`Deleted bundle "${bundleName}".`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to delete bundle.");
    }
  }

  async function handleDeleteWorkspace(workspaceID: string) {
    const workspace = workspaces.find((item) => item.id === workspaceID);
    if (!workspace) {
      return;
    }

    const relatedBundles = bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, workspace.repoPath));
    try {
      setErrorMessage("");
      for (const bundle of relatedBundles) {
        await deleteBundle(bundle.name);
      }

      const nextBundles = bundles.filter((bundle) => !matchesWorkspace(bundle.analyzed_path, workspace.repoPath));
      const nextWorkspaces = workspaces.filter((item) => item.id !== workspaceID);

      setBundleCache((current) => {
        const next = { ...current };
        for (const bundle of relatedBundles) {
          delete next[bundle.name];
        }
        return next;
      });
      setBundles(nextBundles);
      setWorkspaces(nextWorkspaces);

      if (activeWorkspaceID === workspaceID) {
        const nextWorkspace = nextWorkspaces[0] || null;
        setActiveWorkspaceID(nextWorkspace?.id || "");
        setActiveTab("projects");
        if (nextWorkspace) {
          await selectWorkspace(nextWorkspace.id, nextWorkspaces, nextBundles);
        }
      }

      setToastMessage(`Deleted project "${workspace.label}" and its bundles.`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to delete project.");
    }
  }

  const readmeHref = currentBundle ? bundleLink(currentBundle.summary.name, "README.md") : "";

  return (
    <div className="min-h-screen bg-black text-zinc-100">
      <CommandPalette actions={commandActions} onClose={() => setCommandPaletteOpen(false)} open={commandPaletteOpen} />

      <div className="grid min-h-screen xl:grid-cols-[15.5rem_minmax(0,1fr)]">
        <Sidebar
          activeTab={activeTab}
          activeWorkspaceID={activeWorkspaceID}
          onCreateWorkspace={handleCreateWorkspace}
          onSelectTab={setActiveTab}
          onSelectWorkspace={(workspaceID) => {
            void selectWorkspace(workspaceID);
          }}
          workspaces={workspaces}
        />

        <main className="px-4 py-4 xl:px-5">
          <div className="space-y-4">
            <header className="topbar-shell">
              <div>
                <p className="panel-kicker">{tabTitle(activeTab)}</p>
                <h1 className="mt-3 text-2xl font-semibold tracking-tight text-zinc-50">
                  {activeWorkspace?.label || "Codebase Explorer"}
                </h1>
                <p className="mt-2 text-sm text-zinc-500">
                  {activeWorkspace?.repoPath || "Create a project workspace and point it at one local repository."}
                </p>
              </div>
              <div className="flex flex-wrap gap-2">
                <button className="secondary-control" onClick={() => setCommandPaletteOpen(true)} type="button">
                  <Command className="size-4" />
                  Command
                </button>
                <button className="secondary-control" onClick={() => void refreshStatus()} type="button">
                  <RefreshCw className="size-4" />
                  Refresh
                </button>
                <a
                  aria-disabled={!currentBundle}
                  className={!currentBundle ? "secondary-control pointer-events-none opacity-50" : "secondary-control"}
                  href={readmeHref || undefined}
                  rel="noreferrer"
                  target="_blank"
                >
                  <FolderOpenDot className="size-4" />
                  Open README
                </a>
              </div>
            </header>

            {errorMessage ? <div className="error-strip">{errorMessage}</div> : null}
            {status?.bundle_warnings?.length ? (
              <div className="warning-strip">
                {status.bundle_warnings.map((warning) => (
                  <p key={warning}>{warning}</p>
                ))}
              </div>
            ) : null}

            <TabPanels
              activeTab={activeTab}
              activeWorkspace={activeWorkspace}
              activeWorkspaceID={activeWorkspaceID}
              apiKey={apiKey}
              bundle={currentBundle}
              bundles={bundles}
              busy={busy}
              busyDetail={busyDetail}
              inspector={inspector}
              onAPIKeyChange={(value) =>
                setProfileSecrets((current) => ({
                  ...current,
                  [profile.id]: value,
                }))
              }
              onCancelAnalyze={handleCancelAnalyze}
              onCreateWorkspace={handleCreateWorkspace}
              onDeleteBundle={(bundleName) => {
                void handleDeleteBundle(bundleName);
              }}
              onDeleteProfile={handleDeleteProfile}
              onDeleteWorkspace={(workspaceID) => {
                void handleDeleteWorkspace(workspaceID);
              }}
              onDuplicateProfile={handleDuplicateProfile}
              onListProviderModels={() => {
                void handleListProviderModels();
              }}
              onInspect={(value) => {
                setInspector(value);
                if (value) {
                  setActiveTab("properties");
                }
              }}
              onOpenConnections={() => setActiveTab("connections")}
              onProfileChange={(nextProfile) => {
                setProviderDiagnostics(initialProviderDiagnostics);
                setProfiles((current) => current.map((item) => (item.id === nextProfile.id ? nextProfile : item)));
              }}
              onSaveProfile={handleSaveProfile}
              onSelectBundle={(bundleName, workspaceID) => {
                void loadBundle(bundleName, workspaceID);
              }}
              onSelectProfile={(profileID) => {
                setProviderDiagnostics(initialProviderDiagnostics);
                if (activeWorkspace) {
                  updateWorkspace(activeWorkspace.id, { selectedProfile: profileID, updatedAt: new Date().toISOString() });
                }
              }}
              onSelectWorkspace={(workspaceID) => {
                void selectWorkspace(workspaceID);
              }}
              onSubmitAnalyze={handleAnalyze}
              onTestProvider={() => {
                void handleTestProvider();
              }}
              onTabChange={setActiveTab}
              onWorkspaceChange={(workspace) => {
                updateWorkspace(workspace.id, workspace);
              }}
              profile={profile}
              providerDiagnostics={providerDiagnostics}
              profiles={profiles}
              providerOptions={providerOptions}
              repoInputRef={repoInputRef}
              run={activeRun}
              validationErrors={validationErrors}
              workspaces={workspaces}
            />
          </div>
        </main>
      </div>

      {toastMessage ? <div className="toast-shell">{toastMessage}</div> : null}
    </div>
  );
}

function buildCommandActions({
  bundles,
  onAnalyze,
  onCancelAnalyze,
  onCreateWorkspace,
  onFocusAnalyze,
  onRefresh,
  onSelectBundle,
  onSelectTab,
  onSelectWorkspace,
  tabs,
  workspaces,
  busy,
}: {
  bundles: WorkbenchStatusResponse["recent_bundles"];
  onAnalyze: () => void;
  onCancelAnalyze: () => void;
  onCreateWorkspace: () => void;
  onFocusAnalyze: () => void;
  onRefresh: () => void;
  onSelectBundle: (bundleName: string) => void;
  onSelectTab: (tab: TabKey) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  tabs: TabKey[];
  workspaces: SavedWorkspace[];
  busy: boolean;
}): CommandPaletteAction[] {
  const actions: CommandPaletteAction[] = [
    {
      id: "new-project",
      group: "Workspace",
      label: "Create project workspace",
      description: "Create a new saved project shell and focus its repository path input.",
      keywords: ["new", "workspace", "project"],
      run: onCreateWorkspace,
    },
    {
      id: "focus-project-path",
      group: "Workspace",
      label: "Focus repository path",
      description: "Jump directly to the active workspace path input.",
      keywords: ["repo", "path", "focus"],
      shortcut: "A",
      run: onFocusAnalyze,
    },
    {
      id: busy ? "cancel-run" : "run-analyze",
      group: "Analyze",
      label: busy ? "Cancel active analyze run" : "Analyze active project",
      description: busy ? "Stop the current run when it reaches a safe cancel point." : "Run analysis for the active workspace.",
      keywords: ["analyze", "cancel", "run"],
      shortcut: busy ? "Esc" : "Ctrl+Enter",
      run: busy ? onCancelAnalyze : onAnalyze,
    },
    {
      id: "refresh-workbench",
      group: "Workspace",
      label: "Refresh workbench data",
      description: "Refresh status, provider metadata, and available bundles from the local backend.",
      keywords: ["refresh", "reload", "status"],
      shortcut: "R",
      run: onRefresh,
    },
  ];

  for (const tab of tabs) {
    actions.push({
      id: `tab-${tab}`,
      group: "Views",
      label: `Open ${tabTitle(tab)}`,
      description: `Switch the active surface to ${tabTitle(tab)}.`,
      keywords: [tab, "view", "tab"],
      run: () => onSelectTab(tab),
    });
  }

  for (const workspace of workspaces) {
    actions.push({
      id: `workspace-${workspace.id}`,
      group: "Projects",
      label: `Open ${workspace.label}`,
      description: workspace.repoPath || "Saved workspace without a repository path yet.",
      keywords: [workspace.label, workspace.repoPath].filter(Boolean),
      run: () => onSelectWorkspace(workspace.id),
    });
  }

  for (const bundle of bundles) {
    actions.push({
      id: `bundle-${bundle.name}`,
      group: "Bundles",
      label: `Open ${bundle.project_name || bundle.name}`,
      description: `${bundle.total_files} files · AI ${bundle.ai_status || "disabled"}`,
      keywords: [bundle.name, bundle.project_name, bundle.ai_status].filter(Boolean) as string[],
      run: () => onSelectBundle(bundle.name),
    });
  }

  return actions;
}

function buildDefaultInspector(workspace: SavedWorkspace | null, bundle: WorkbenchBundle | null, activeTab: TabKey): InspectorState | null {
  if (bundle) {
    return {
      eyebrow: tabTitle(activeTab),
      title: bundle.summary.project_name || bundle.summary.name,
      description: bundle.data.project.summary || bundle.data.ai.project_summary || "No project summary available.",
      notes: [bundle.data.ai.note || "Repository bundle with optional AI synthesis."],
      properties: [
        { label: "Workspace", value: workspace?.label || "Unknown" },
        { label: "Bundle", value: bundle.summary.name },
        { label: "Files", value: String(bundle.summary.total_files) },
        { label: "Lines", value: String(bundle.summary.total_lines) },
      ],
    };
  }

  if (!workspace) {
    return null;
  }

  return {
    eyebrow: "Workspace",
    title: workspace.label,
    description: workspace.repoPath || "Set a local repository path to begin.",
    notes: ["One active project per app instance. Saved locally on this machine."],
    properties: [
      { label: "Support files", value: String(workspace.supportFiles.length) },
      { label: "Ignore patterns", value: String(workspace.ignorePatterns.length) },
      { label: "Connection", value: workspace.selectedProfile || "unassigned" },
    ],
  };
}

function matchesWorkspace(bundlePath: string, workspacePath: string) {
  if (!workspacePath.trim()) {
    return false;
  }

  return normalizeLocalPath(bundlePath) === normalizeLocalPath(workspacePath);
}

function sortWorkspacesByUpdatedAt(left: SavedWorkspace, right: SavedWorkspace) {
  return right.updatedAt.localeCompare(left.updatedAt);
}

function tabTitle(tab: TabKey) {
  switch (tab) {
    case "projects":
      return "Projects";
    case "dashboard":
      return "Dashboard";
    case "project":
      return "Project Setup";
    case "connections":
      return "Connections";
    case "properties":
      return "Properties";
    case "summary":
      return "Summary";
    case "architecture":
      return "Architecture";
    case "flowchart":
      return "Flowchart";
    case "issues":
      return "Issues";
    case "recommendations":
      return "Recommendations";
    default:
      return "Workspace";
  }
}
