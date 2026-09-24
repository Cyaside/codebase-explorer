import { startTransition, useEffect, useRef, useState } from "react";
import { FolderOpenDot, Play, RefreshCw } from "lucide-react";

import { Sidebar } from "@/components/Sidebar";
import { TabPanels } from "@/components/TabPanels";
import {
  buildProviderPayload,
  cancelAnalyzeRun,
  deleteBundle,
  deleteSavedCredential,
  fetchAnalyzeRun,
  fetchBundle,
  fetchProviderModels,
  fetchSavedCredentials,
  fetchStatus,
  mergeBundleSummary,
  saveCredential,
  startAnalyzeRun,
  testProvider,
} from "@/lib/api";
import { createWorkspace, loadProfile, loadUIState, loadWorkspaces, persistProfile, persistUIState, persistWorkspaces } from "@/lib/storage";
import { openAnalyzeRunStream } from "@/lib/stream";
import type {
  AnalyzeRun,
  InspectorState,
  ProviderDiagnosticsState,
  SavedWorkspace,
  TabKey,
  WorkbenchBundle,
  WorkbenchStatusResponse,
} from "@/lib/types";
import { bundleLink, matchesWorkspace } from "@/lib/utils";
import { validateProfile } from "@/lib/validation";

const initialUIState = loadUIState();
const initialProfile = loadProfile();
const initialWorkspaces = loadWorkspaces();
const ACTIVE_RUN_STORAGE_KEY = "codearch.workbench.active-run.v1";
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
  const [profile, setProfile] = useState(initialProfile);
  const [workspaces, setWorkspaces] = useState(initialWorkspaces);
  const [activeWorkspaceID, setActiveWorkspaceID] = useState(initialUIState.activeWorkspace || initialWorkspaces[0]?.id || "");
  const [activeTab, setActiveTab] = useState<TabKey>(initialUIState.activeTab || "projects");
  const [apiKey, setAPIKey] = useState("");
  const [savedKeyBaseURL, setSavedKeyBaseURL] = useState("");
  const [activeRun, setActiveRun] = useState<AnalyzeRun | null>(null);
  const [busyDetail, setBusyDetail] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [toastMessage, setToastMessage] = useState("");
  const [inspector, setInspector] = useState<InspectorState | null>(null);
  const [providerDiagnostics, setProviderDiagnostics] = useState<ProviderDiagnosticsState>(initialProviderDiagnostics);

  const repoInputRef = useRef<HTMLInputElement | null>(null);
  const bundleRequestsRef = useRef(new Map<string, Promise<WorkbenchBundle>>());
  const completedRunRef = useRef("");
  const runWorkspaceIDRef = useRef("");

  const activeWorkspace = workspaces.find((workspace) => workspace.id === activeWorkspaceID) || null;
  const providerOptions = status?.supported_providers || [];
  const hasSavedKey = !!savedKeyBaseURL && savedKeyBaseURL === profile.provider.baseUrl.trim().replace(/\/+$/, "");
  const validationErrors = validateProfile(profile, apiKey, providerOptions, hasSavedKey);
  const workspaceBundles = activeWorkspace ? bundles.filter((bundle) => matchesWorkspace(bundle.analyzed_path, activeWorkspace.repoPath)) : [];
  const selectedBundleName = activeWorkspace?.activeBundle || workspaceBundles[0]?.name || "";
  const currentBundle = selectedBundleName ? bundleCache[selectedBundleName] || null : null;
  const busy = !!activeRun && ["queued", "running", "canceling"].includes(activeRun.status);

  useEffect(() => {
    void refreshStatus();
    void fetchSavedCredentials().then((connections) => {
      const saved = connections.find((item) => item.id === "openai");
      if (saved) {
        setSavedKeyBaseURL(saved.base_url);
        setProfile({ id: "openai", label: "OpenAI-compatible", provider: { name: "compatible", model: saved.model, baseUrl: saved.base_url } });
      }
    }).catch((error) => setErrorMessage(error instanceof Error ? error.message : "Could not load saved connections."));
    try {
      const persisted = JSON.parse(window.sessionStorage.getItem(ACTIVE_RUN_STORAGE_KEY) || "null") as { id?: string; workspace_id?: string } | null;
      if (persisted?.id) {
        runWorkspaceIDRef.current = persisted.workspace_id || "";
        void fetchAnalyzeRun(persisted.id).then((response) => applyRunSnapshot(response.run)).catch(() => {
          window.sessionStorage.removeItem(ACTIVE_RUN_STORAGE_KEY);
        });
      }
    } catch {
      window.sessionStorage.removeItem(ACTIVE_RUN_STORAGE_KEY);
    }
  }, []);

  useEffect(() => {
    persistProfile(profile);
  }, [profile]);

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

      if (activeWorkspace && nextBundleName) {
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
        let request = bundleRequestsRef.current.get(bundleName);
        if (!request) {
          request = fetchBundle(bundleName);
          bundleRequestsRef.current.set(bundleName, request);
        }
        let bundle: WorkbenchBundle;
        try {
          bundle = await request;
        } finally {
          bundleRequestsRef.current.delete(bundleName);
        }
        setBundleCache((current) => {
          const retained = Object.entries(current).filter(([name]) => name !== bundleName);
          const selected = retained.find(([name]) => name === selectedBundleName);
          const recent = retained.filter(([name]) => name !== selectedBundleName).slice(selected ? -1 : -2);
          return Object.fromEntries([...recent, ...(selected ? [selected] : []), [bundleName, bundle]]);
        });
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
    if (["succeeded", "partial", "failed", "canceled"].includes(run.status)) {
      window.sessionStorage.removeItem(ACTIVE_RUN_STORAGE_KEY);
    }

    if ((run.status === "succeeded" || run.status === "partial") && run.response) {
      if (completedRunRef.current === run.id) {
        return;
      }
      completedRunRef.current = run.id;
      const response = run.response;
      setErrorMessage("");
      setBundles((current) => mergeBundleSummary(current, response.bundle));

      const runWorkspaceID = runWorkspaceIDRef.current || activeWorkspace?.id || "";
      const runWorkspace = workspaces.find((workspace) => workspace.id === runWorkspaceID);
      if (runWorkspaceID) {
        updateWorkspace(runWorkspaceID, {
          activeBundle: response.bundle.name,
          label: runWorkspace?.label === "Untitled Project" ? response.bundle.project_name || runWorkspace.label : runWorkspace?.label || "Untitled Project",
          updatedAt: new Date().toISOString(),
        });
      }

      setToastMessage(run.status === "partial"
        ? `Partial analysis for ${response.bundle.project_name || response.result.project_name || "repository"}; some AI sections failed.`
        : `Analysis ready for ${response.bundle.project_name || response.result.project_name || "repository"}.`);
      if (runWorkspaceID && runWorkspaceID !== activeWorkspace?.id) {
        void loadBundle(response.bundle.name, runWorkspaceID);
        void refreshStatus();
      } else {
        void refreshStatus(response.bundle.name);
      }
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

  async function currentConnectionPayload() {
    const saved = await saveCredential(profile, apiKey.trim());
    setSavedKeyBaseURL(saved.base_url);
    if (apiKey.trim()) setAPIKey("");
    return { provider: buildProviderPayload(profile, ""), credential_id: profile.id };
  }

  async function handleSaveProfile() {
    if (validationErrors.length) {
      setErrorMessage(validationErrors.join(" "));
      return;
    }
    try {
      await currentConnectionPayload();
      setErrorMessage("");
      setToastMessage(`Saved connection "${profile.label}" on this computer.`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not save connection.");
    }
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
      const response = await fetchProviderModels(await currentConnectionPayload());
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
      const response = await testProvider(await currentConnectionPayload());
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

  async function handleClearSavedKey() {
    try {
      await deleteSavedCredential(profile.id);
      setSavedKeyBaseURL("");
      setToastMessage(`Cleared the saved key for "${profile.label}".`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Could not clear saved key.");
    }
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
      const connection = await currentConnectionPayload();
      const response = await startAnalyzeRun({
        repo_path: repoPath,
        support_files: activeWorkspace.supportFiles,
        extra_ignore_patterns: activeWorkspace.ignorePatterns,
        ...connection,
      });
      runWorkspaceIDRef.current = activeWorkspace.id;
      window.sessionStorage.setItem(ACTIVE_RUN_STORAGE_KEY, JSON.stringify({ id: response.run.id, workspace_id: activeWorkspace.id }));
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
    <div className="app-shell">
      <a className="skip-link" href="#workbench-main">Skip to content</a>

      <div className="app-layout">
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

        <main className="workbench-main" id="workbench-main" tabIndex={-1}>
          <div className="workbench-content">
            <header className="workbench-header">
              <div className="workbench-breadcrumb">
                <span>Workspace</span><span aria-hidden="true">/</span>
                <span>{activeWorkspace?.label || "No project"}</span><span aria-hidden="true">/</span>
                <strong>{tabTitle(activeTab)}</strong>
              </div>
              <div className="workbench-heading">
                <div className="min-w-0">
                  <p className="panel-kicker">{currentBundle ? `${currentBundle.summary.project_type || "Repository"} · ${currentBundle.summary.total_files} files` : "Local workbench"}</p>
                  <h1>{tabTitle(activeTab)}</h1>
                  <p className="workbench-subtitle">{activeWorkspace?.repoPath || "Select a project and connect a repository to begin."}</p>
                </div>
                <div className="workbench-actions">
                  <button aria-label="Refresh workbench" className="secondary-control icon-control" onClick={() => void refreshStatus()} title="Refresh" type="button"><RefreshCw className="size-4" /></button>
                  {currentBundle ? <a className="secondary-control icon-control" href={readmeHref} rel="noreferrer" target="_blank" title="Open bundle README" aria-label="Open bundle README"><FolderOpenDot className="size-4" /></a> : null}
                  {activeTab !== "project" ? <button className="primary-control" onClick={() => setActiveTab("project")} type="button"><Play className="size-3.5" /> Analyze</button> : null}
                </div>
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
              hasSavedKey={hasSavedKey}
              bundle={currentBundle}
              bundles={bundles}
              busy={busy}
              busyDetail={busyDetail}
              inspector={inspector}
              onAPIKeyChange={setAPIKey}
              onCancelAnalyze={handleCancelAnalyze}
              onClearSavedKey={() => { void handleClearSavedKey(); }}
              onCreateWorkspace={handleCreateWorkspace}
              onDeleteBundle={(bundleName) => {
                void handleDeleteBundle(bundleName);
              }}
              onDeleteWorkspace={(workspaceID) => {
                void handleDeleteWorkspace(workspaceID);
              }}
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
                setProfile(nextProfile);
              }}
              onSaveProfile={() => { void handleSaveProfile(); }}
              onSelectBundle={(bundleName, workspaceID) => {
                void loadBundle(bundleName, workspaceID);
              }}
              onSelectWorkspace={(workspaceID) => {
                void selectWorkspace(workspaceID);
              }}
              onSubmitAnalyze={handleAnalyze}
              onTestProvider={() => {
                void handleTestProvider();
              }}
              onWorkspaceChange={(workspace) => {
                updateWorkspace(workspace.id, workspace);
              }}
              profile={profile}
              providerDiagnostics={providerDiagnostics}
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

function buildDefaultInspector(workspace: SavedWorkspace | null, bundle: WorkbenchBundle | null, activeTab: TabKey): InspectorState | null {
  if (bundle) {
    return {
      eyebrow: tabTitle(activeTab),
      title: bundle.summary.project_name || bundle.summary.name,
      description: bundle.data.project.summary || bundle.data.ai.project_summary || "No project summary available.",
      notes: [bundle.data.ai.note || "Repository analysis with provider synthesis."],
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
      { label: "Connection", value: "OpenAI-compatible" },
    ],
  };
}

function sortWorkspacesByUpdatedAt(left: SavedWorkspace, right: SavedWorkspace) {
  return right.updatedAt.localeCompare(left.updatedAt);
}

function tabTitle(tab: TabKey) {
  switch (tab) {
    case "projects":
      return "Projects";
    case "dashboard":
      return "Overview";
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
      return "Graphs";
    case "issues":
      return "Issues";
    case "recommendations":
      return "Recommendations";
    default:
      return "Workspace";
  }
}
