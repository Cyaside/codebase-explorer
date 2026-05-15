import type { RefObject } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import { FolderCog } from "lucide-react";

import { ArchitectureView, DashboardView, IssuesView, RecommendationsView, SummaryView } from "@/components/AnalysisViews";
import { AnalyzeForm } from "@/components/AnalyzeForm";
import { ConnectionPanel } from "@/components/ConnectionPanel";
import { EvidenceGraph } from "@/components/EvidenceGraph";
import { ProjectsPanel } from "@/components/ProjectsPanel";
import { PropertiesPanel } from "@/components/PropertiesPanel";
import type {
  AnalyzeRun,
  BundleSummary,
  ConnectionProfile,
  InspectorState,
  ProviderDiagnosticsState,
  SavedWorkspace,
  SupportedProviderOption,
  TabKey,
  WorkbenchBundle,
} from "@/lib/types";

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
  onDeleteProfile: () => void;
  onDeleteWorkspace: (workspaceID: string) => void;
  onDuplicateProfile: () => void;
  onInspect: (value: InspectorState | null) => void;
  onListProviderModels: () => void;
  onOpenConnections: () => void;
  onProfileChange: (profile: ConnectionProfile) => void;
  onSaveProfile: () => void;
  onSelectBundle: (bundleName: string, workspaceID?: string) => void;
  onSelectProfile: (profileID: string) => void;
  onSelectWorkspace: (workspaceID: string) => void;
  onSubmitAnalyze: () => void;
  onTabChange: (value: TabKey) => void;
  onTestProvider: () => void;
  onWorkspaceChange: (workspace: SavedWorkspace) => void;
  profile: ConnectionProfile;
  providerDiagnostics: ProviderDiagnosticsState;
  profiles: ConnectionProfile[];
  providerOptions: SupportedProviderOption[];
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
    onDeleteProfile,
    onDeleteWorkspace,
    onDuplicateProfile,
    onInspect,
    onListProviderModels,
    onOpenConnections,
    onProfileChange,
    onSaveProfile,
    onSelectBundle,
    onSelectProfile,
    onSelectWorkspace,
    onSubmitAnalyze,
    onTabChange,
    onTestProvider,
    onWorkspaceChange,
    profile,
    providerDiagnostics,
    profiles,
    providerOptions,
    repoInputRef,
    run,
    validationErrors,
    workspaces,
  } = props;

  return (
    <Tabs.Root className="space-y-4" onValueChange={(value) => onTabChange(value as TabKey)} value={activeTab}>
      <Tabs.Content value="projects">
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
      </Tabs.Content>

      <Tabs.Content value="project">
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
      </Tabs.Content>

      <Tabs.Content value="connections">
        <div className="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
          <section className="panel-block">
            <div className="flex items-center gap-3">
              <FolderCog className="size-4 text-zinc-500" />
              <p className="panel-kicker">Connection profiles</p>
            </div>
            <div className="mt-4 space-y-2">
              {profiles.map((item) => (
                <button
                  className={`panel-row${item.id === profile.id ? " panel-row-selected" : ""}`}
                  key={item.id}
                  onClick={() => onSelectProfile(item.id)}
                  type="button"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-zinc-100">{item.label}</p>
                    <p className="mt-1 truncate text-xs text-zinc-500">
                      {item.provider.name} - {item.provider.model}
                    </p>
                  </div>
                  <span className="text-xs text-zinc-600">{item.id === activeWorkspace?.selectedProfile ? "In use" : ""}</span>
                </button>
              ))}
            </div>
          </section>

          <ConnectionPanel
            apiKey={apiKey}
            hasSavedKey={hasSavedKey}
            diagnostics={providerDiagnostics}
            onAPIKeyChange={onAPIKeyChange}
            onChange={onProfileChange}
            onClearKey={onClearSavedKey}
            onDelete={onDeleteProfile}
            onDuplicate={onDuplicateProfile}
            onListModels={onListProviderModels}
            onSave={onSaveProfile}
            onTestProvider={onTestProvider}
            profile={profile}
            providerOptions={providerOptions}
            validationErrors={validationErrors}
          />
        </div>
      </Tabs.Content>

      <Tabs.Content value="properties">
        <PropertiesPanel activeRun={run} bundle={bundle} inspector={inspector} workspace={activeWorkspace} />
      </Tabs.Content>

      <Tabs.Content value="dashboard">
        <DashboardView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="summary">
        <SummaryView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="architecture">
        <ArchitectureView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="flowchart">
        <div className="space-y-4">
          <EvidenceGraph bundle={bundle} />
        </div>
      </Tabs.Content>

      <Tabs.Content value="issues">
        <IssuesView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>

      <Tabs.Content value="recommendations">
        <RecommendationsView bundle={bundle} onInspect={onInspect} />
      </Tabs.Content>
    </Tabs.Root>
  );
}
