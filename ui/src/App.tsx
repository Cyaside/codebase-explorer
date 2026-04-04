import { useEffect, useRef, useState } from "react";
import { CircleAlert, FolderOpenDot } from "lucide-react";

import { AnalyzeForm } from "@/components/AnalyzeForm";
import { ConnectionPanel } from "@/components/ConnectionPanel";
import { Sidebar } from "@/components/Sidebar";
import { StatsGrid } from "@/components/StatsGrid";
import { TabPanels } from "@/components/TabPanels";
import { WorkbenchAlerts } from "@/components/WorkbenchAlerts";
import { WorkbenchRail } from "@/components/WorkbenchRail";
import { buildProviderPayload, cancelAnalyzeRun, fetchAnalyzeRun, fetchBundle, fetchStatus, mergeBundleSummary, startAnalyzeRun } from "@/lib/api";
import { useWorkbenchShortcuts } from "@/hooks/useWorkbenchShortcuts";
import { defaultProfiles, loadProfiles, loadUIState, persistProfiles, persistUIState } from "@/lib/storage";
import { openAnalyzeRunStream } from "@/lib/stream";
import type { AnalyzeFormState, AnalyzeRun, ConnectionProfile, TabKey, WorkbenchBundle, WorkbenchStatusResponse } from "@/lib/types";
import { bundleLink, splitLines } from "@/lib/utils";
import { validateProfile } from "@/lib/validation";

const initialUIState = loadUIState();
const initialProfiles = loadProfiles();
const workbenchTabs: TabKey[] = ["summary", "architecture", "flowchart", "issues", "recommendations"];

export function App() {
  const [status, setStatus] = useState<WorkbenchStatusResponse | null>(null);
  const [bundles, setBundles] = useState<WorkbenchStatusResponse["recent_bundles"]>([]);
  const [bundleCache, setBundleCache] = useState<Record<string, WorkbenchBundle>>({});
  const [selectedBundle, setSelectedBundle] = useState(initialUIState.selectedBundle);
  const [profiles, setProfiles] = useState(initialProfiles);
  const [selectedProfile, setSelectedProfile] = useState(initialUIState.selectedProfile || initialProfiles[0]?.id || "deterministic");
  const [activeTab, setActiveTab] = useState<TabKey>(initialUIState.activeTab || "summary");
  const [form, setForm] = useState<AnalyzeFormState>({
    repoPath: "",
    supportFiles: "",
    ignorePatterns: "",
  });
  const [profileSecrets, setProfileSecrets] = useState<Record<string, string>>({});
  const [activeRun, setActiveRun] = useState<AnalyzeRun | null>(null);
  const [busyDetail, setBusyDetail] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [toastMessage, setToastMessage] = useState("");

  const profile = profiles.find((item) => item.id === selectedProfile) || profiles[0] || defaultProfiles()[0];
  const currentBundle = selectedBundle ? bundleCache[selectedBundle] || null : null;
  const providerOptions = status?.supported_providers || [];
  const apiKey = profileSecrets[profile.id] || "";
  const validationErrors = validateProfile(profile, apiKey, providerOptions);
  const busy = !!activeRun && ["queued", "running", "canceling"].includes(activeRun.status);
  const repoInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    void refreshStatus();
  }, []);

  useEffect(() => {
    persistProfiles(profiles);
  }, [profiles]);

  useEffect(() => {
    persistUIState({
      activeTab,
      selectedBundle,
      selectedProfile,
    });
  }, [activeTab, selectedBundle, selectedProfile]);

  useEffect(() => {
    if (!toastMessage) {
      return;
    }

    const timeoutID = window.setTimeout(() => {
      setToastMessage("");
    }, 3200);

    return () => window.clearTimeout(timeoutID);
  }, [toastMessage]);

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

      intervalID = window.setInterval(async () => {
        await pollOnce();
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

  useWorkbenchShortcuts({
    activeTab,
    busy,
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
    onFocusAnalyze: () => {
      repoInputRef.current?.focus();
      repoInputRef.current?.select();
    },
    onRefresh: () => {
      void refreshStatus();
    },
    onSelectTab: setActiveTab,
    tabs: workbenchTabs,
  });

  async function refreshStatus(preferredBundleName?: string) {
    try {
      setErrorMessage("");
      const nextStatus = await fetchStatus();
      setStatus(nextStatus);
      setBundles(nextStatus.recent_bundles || []);

      const requestedBundle = preferredBundleName || selectedBundle;
      const nextSelected =
        requestedBundle && nextStatus.recent_bundles.some((bundle) => bundle.name === requestedBundle)
          ? requestedBundle
          : nextStatus.recent_bundles[0]?.name || "";

      if (!nextSelected) {
        setSelectedBundle("");
        return;
      }

      await loadBundle(nextSelected);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to load workbench status.");
    }
  }

  async function loadBundle(bundleName: string) {
    try {
      setErrorMessage("");
      if (!bundleCache[bundleName]) {
        const bundle = await fetchBundle(bundleName);
        setBundleCache((current) => ({
          ...current,
          [bundleName]: bundle,
        }));
      }
      setSelectedBundle(bundleName);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "Failed to load bundle.");
    }
  }

  async function selectRelativeBundle(direction: -1 | 1) {
    if (!bundles.length) {
      return;
    }

    const currentIndex = bundles.findIndex((bundle) => bundle.name === selectedBundle);
    const startIndex = currentIndex >= 0 ? currentIndex : 0;
    const nextIndex = (startIndex + direction + bundles.length) % bundles.length;
    await loadBundle(bundles[nextIndex].name);
  }

  function applyRunSnapshot(run: AnalyzeRun) {
    setActiveRun(run);
    const latestEvent = run.progress[run.progress.length - 1];
    setBusyDetail(latestEvent?.detail || run.status);

    const response = run.response;
    if (run.status === "succeeded" && response) {
      setErrorMessage("");
      setBundleCache((current) => ({
        ...current,
        [response.bundle.name]: {
          summary: response.bundle,
          data: response.data,
        },
      }));
      setBundles((current) => mergeBundleSummary(current, response.bundle));
      setSelectedBundle(response.bundle.name);
      setToastMessage(`Analysis ready for ${response.bundle.project_name || response.result.project_name || "repository"}.`);
      void refreshStatus(response.bundle.name);
      return;
    }

    if (run.status === "failed") {
      setErrorMessage(run.error || "Analyze run failed.");
      return;
    }

    if (run.status === "canceled") {
      setErrorMessage("");
      setToastMessage("Analyze run canceled.");
    }
  }

  function updateCurrentProfile(nextProfile: ConnectionProfile) {
    setProfiles((current) =>
      current.map((item) => {
        if (item.id !== nextProfile.id) {
          return item;
        }
        return nextProfile;
      }),
    );
  }

  function handleSaveProfile() {
    if (validationErrors.length) {
      setErrorMessage(validationErrors.join(" "));
      return;
    }

    setErrorMessage("");
    setToastMessage(`Saved connection "${profile.label}". API key stays only in memory for this browser session.`);
  }

  function handleDuplicateProfile() {
    const cloneID = `profile-${Date.now()}`;
    const clone: ConnectionProfile = {
      id: cloneID,
      label: `${profile.label} Copy`,
      provider: profile.provider ? { ...profile.provider } : null,
    };

    setProfiles((current) => [...current, clone]);
    setSelectedProfile(cloneID);
    setProfileSecrets((current) => ({
      ...current,
      [cloneID]: current[profile.id] || "",
    }));
    setToastMessage(`Duplicated "${profile.label}".`);
  }

  function handleDeleteProfile() {
    if (profile.locked) {
      setErrorMessage("Preset profiles cannot be deleted.");
      return;
    }

    const remaining = profiles.filter((item) => item.id !== profile.id);
    setProfiles(remaining);
    setSelectedProfile(remaining[0]?.id || "deterministic");
    setProfileSecrets((current) => {
      const next = { ...current };
      delete next[profile.id];
      return next;
    });
    setToastMessage(`Deleted "${profile.label}".`);
  }

  async function handleAnalyze() {
    const repoPath = form.repoPath.trim();
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
        deterministic_only: !profile.provider,
        support_files: splitLines(form.supportFiles),
        extra_ignore_patterns: splitLines(form.ignorePatterns),
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

  const readmeHref = currentBundle ? bundleLink(currentBundle.summary.name, "README.md") : "";

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="grid min-h-screen lg:grid-cols-[17.5rem_minmax(0,1fr)]">
        <Sidebar
          bundles={bundles}
          onRefresh={() => {
            void refreshStatus();
          }}
          onSelectBundle={(bundleName) => {
            void loadBundle(bundleName);
          }}
          onSelectProfile={setSelectedProfile}
          profiles={profiles}
          selectedBundle={selectedBundle}
          selectedProfile={selectedProfile}
          status={status}
        />

        <main className="space-y-5 px-4 py-5 lg:px-6">
          <div className="grid gap-5 2xl:grid-cols-[minmax(0,1fr)_21rem]">
            <div className="space-y-5">
              <header className="rounded-3xl border border-border bg-card/90 p-5 shadow-sm">
                <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                  <div>
                    <p className="text-[10px] uppercase tracking-[0.28em] text-primary">Local-First Repository Orientation</p>
                    <h1 className="mt-2 text-2xl font-bold">
                      {currentBundle ? currentBundle.summary.project_name || "Workbench" : "Codebase Explorer Workbench"}
                    </h1>
                    <p className="mt-3 text-sm text-muted-foreground">
                      {currentBundle
                        ? `${currentBundle.summary.project_type || "Repository"} · ${currentBundle.summary.total_files} files · ${currentBundle.summary.total_lines} lines`
                        : "Open a local repository, run analysis, and navigate reusable orientation bundles without leaving your machine."}
                    </p>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <button className="secondary-button" onClick={() => void refreshStatus()} type="button">
                      <FolderOpenDot className="size-4" />
                      Reload Dashboard
                    </button>
                    <a
                      aria-disabled={!currentBundle}
                      className={!currentBundle ? "secondary-button pointer-events-none opacity-50" : "secondary-button"}
                      href={readmeHref || undefined}
                      rel="noreferrer"
                      target="_blank"
                    >
                      Open Bundle README
                    </a>
                  </div>
                </div>
              </header>

              <WorkbenchAlerts activeRun={activeRun} bundle={currentBundle} errorMessage={errorMessage} status={status} />

              <section className="grid gap-4 2xl:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)]">
                <AnalyzeForm
                  activeProfileLabel={profile.label}
                  busy={busy}
                  busyDetail={busyDetail}
                  form={form}
                  onCancel={handleCancelAnalyze}
                  onChange={setForm}
                  onSubmit={handleAnalyze}
                  repoInputRef={repoInputRef}
                  run={activeRun}
                />
                <ConnectionPanel
                  apiKey={apiKey}
                  onAPIKeyChange={(value) =>
                    setProfileSecrets((current) => ({
                      ...current,
                      [profile.id]: value,
                    }))
                  }
                  onChange={updateCurrentProfile}
                  onDelete={handleDeleteProfile}
                  onDuplicate={handleDuplicateProfile}
                  onSave={handleSaveProfile}
                  profile={profile}
                  providerOptions={providerOptions}
                  validationErrors={validationErrors}
                />
              </section>

              <StatsGrid bundle={currentBundle} bundles={bundles} connections={profiles.length} status={status} />

              <TabPanels activeTab={activeTab} bundle={currentBundle} onTabChange={setActiveTab} />
            </div>

            <WorkbenchRail activeRun={activeRun} bundle={currentBundle} />
          </div>
        </main>
      </div>

      {toastMessage ? (
        <div className="pointer-events-none fixed bottom-5 right-5 z-50 flex items-center gap-3 rounded-2xl border border-border bg-card px-4 py-3 text-sm text-foreground shadow-lg">
          <CircleAlert className="size-4 text-primary" />
          <span>{toastMessage}</span>
        </div>
      ) : null}
    </div>
  );
}
