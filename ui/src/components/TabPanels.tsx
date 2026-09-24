import { lazy, Suspense, type RefObject } from "react";

import { ArchitectureView, DashboardView, IssuesView, RecommendationsView, SummaryView } from "@/components/AnalysisViews";
import { AnalyzeForm } from "@/components/AnalyzeForm";
import { ConnectionPanel } from "@/components/ConnectionPanel";
import { ProjectsPanel } from "@/components/ProjectsPanel";
import { PropertiesPanel } from "@/components/PropertiesPanel";
import type {
  AnalyzeRun,
  BundleSummary,
  ConnectionProfile,
  InspectorState,
  ProviderDiagnosticsState,
  SavedWorkspace,
  TabKey,
  WorkbenchBundle,
} from "@/lib/types";

const EvidenceGraph = lazy(() => import("@/components/EvidenceGraph").then(({ EvidenceGraph }) => ({ default: EvidenceGraph })));

interface TabPanelsProps {
  activeTab: TabKey;
  activeWorkspace: SavedWorkspace | null;
  activeWorkspaceID: string;
  apiKey: string;
  hasSavedKey: boolean;
  bundle: WorkbenchBundle | null;
  bundles: BundleSummary[];
  busy: boolean;
  busyDetail: string;
  inspector: InspectorState | null;
  onAPIKeyChange: (value: string) => void;
  onCancelAnalyze: () => void;
  onClearSavedKey: () => void;
  onCreateWorkspace: () => void;
  onDeleteBundle: (bundleName: string) => void;
  onDeleteWorkspace: (workspaceID: string) => void;
  onInspect: (value: InspectorState | null) => void;
  onListProviderModels: () => void;
  onOpenConnections: () => void;
  onProfileChange: (profile: ConnectionProfile) => void;
  onSaveProfile: () => void;
  onSelectBundle: (bundleName: string, workspaceID?: string) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  onSubmitAnalyze: () => void;
  onTestProvider: () => void;
  onWorkspaceChange: (workspace: SavedWorkspace) => void;
  profile: ConnectionProfile;
  providerDiagnostics: ProviderDiagnosticsState;
  repoInputRef: RefObject<HTMLInputElement | null>;
  run: AnalyzeRun | null;
  validationErrors: string[];
  workspaces: SavedWorkspace[];
}

export function TabPanels(props: TabPanelsProps) {
  const {
    activeTab,
    activeWorkspace,
    activeWorkspaceID,
    apiKey,
    hasSavedKey,
    bundle,
    bundles,
    busy,
    busyDetail,
    inspector,
    onAPIKeyChange,
    onCancelAnalyze,
    onClearSavedKey,
    onCreateWorkspace,
    onDeleteBundle,
    onDeleteWorkspace,
    onInspect,
    onListProviderModels,
    onOpenConnections,
    onProfileChange,
    onSaveProfile,
    onSelectBundle,
    onSelectWorkspace,
    onSubmitAnalyze,
    onTestProvider,
    onWorkspaceChange,
    profile,
    providerDiagnostics,
    repoInputRef,
    run,
    validationErrors,
    workspaces,
  } = props;

  return (
    <div className="space-y-4">
      {activeTab === "projects" ? (
        <ProjectsPanel
          activeWorkspaceID={activeWorkspaceID}
          bundles={bundles}
          onCreateWorkspace={onCreateWorkspace}
          onDeleteBundle={onDeleteBundle}
          onDeleteWorkspace={onDeleteWorkspace}
          onSelectBundle={onSelectBundle}
          onSelectWorkspace={onSelectWorkspace}
          selectedBundle={activeWorkspace?.activeBundle || ""}
          workspaces={workspaces}
        />
      ) : null}

      {activeTab === "project" ? (
        <AnalyzeForm
          busy={busy}
          busyDetail={busyDetail}
          connectionReady={validationErrors.length === 0}
          currentConnectionLabel={profile.label}
          onCancel={onCancelAnalyze}
          onCreateWorkspace={onCreateWorkspace}
          onOpenConnections={onOpenConnections}
          onSubmit={onSubmitAnalyze}
          onWorkspaceChange={onWorkspaceChange}
          repoInputRef={repoInputRef}
          run={run}
          workspace={activeWorkspace}
        />
      ) : null}

      {activeTab === "connections" ? (
        <div className="mx-auto max-w-3xl">
          <ConnectionPanel
            apiKey={apiKey}
            hasSavedKey={hasSavedKey}
            diagnostics={providerDiagnostics}
            onAPIKeyChange={onAPIKeyChange}
            onChange={onProfileChange}
            onClearKey={onClearSavedKey}
            onListModels={onListProviderModels}
            onSave={onSaveProfile}
            onTestProvider={onTestProvider}
            profile={profile}
            validationErrors={validationErrors}
          />
        </div>
      ) : null}

      {activeTab === "properties" ? (
        <PropertiesPanel activeRun={run} bundle={bundle} inspector={inspector} workspace={activeWorkspace} />
      ) : null}

      {activeTab === "dashboard" ? (
        <DashboardView bundle={bundle} onInspect={onInspect} />
      ) : null}

      {activeTab === "summary" ? (
        <SummaryView bundle={bundle} onInspect={onInspect} />
      ) : null}

      {activeTab === "architecture" ? (
        <ArchitectureView bundle={bundle} onInspect={onInspect} />
      ) : null}

      {activeTab === "flowchart" ? (
        <Suspense fallback={<div className="panel-block text-sm text-zinc-400">Loading graph…</div>}>
          <EvidenceGraph bundle={bundle} />
        </Suspense>
      ) : null}

      {activeTab === "issues" ? (
        <IssuesView bundle={bundle} onInspect={onInspect} />
      ) : null}

      {activeTab === "recommendations" ? (
        <RecommendationsView bundle={bundle} onInspect={onInspect} />
      ) : null}
    </div>
  );
}
