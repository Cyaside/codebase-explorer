import type { ReactNode, RefObject } from "react";
import * as Tabs from "@radix-ui/react-tabs";
import { FolderCog } from "lucide-react";

import { ArchitectureView, DashboardView, IssuesView, RecommendationsView, SummaryView } from "@/components/AnalysisViews";
import { AnalyzeForm } from "@/components/AnalyzeForm";
import { ConnectionPanel } from "@/components/ConnectionPanel";
import { FlowchartCanvas } from "@/components/FlowchartCanvas";
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
  bundle: WorkbenchBundle | null;
  bundles: BundleSummary[];
  busy: boolean;
  busyDetail: string;
  inspector: InspectorState | null;
  onAPIKeyChange: (value: string) => void;
  onCancelAnalyze: () => void;
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
    bundle,
    bundles,
    busy,
    busyDetail,
    inspector,
    onAPIKeyChange,
    onCancelAnalyze,
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
      <Tabs.List className="tab-strip">
        <TabTrigger value="projects">Projects</TabTrigger>
        <TabTrigger value="project">Project Setup</TabTrigger>
        <TabTrigger value="connections">Connections</TabTrigger>
        <TabTrigger value="properties">Properties</TabTrigger>
        <TabTrigger value="dashboard">Dashboard</TabTrigger>
        <TabTrigger value="summary">Summary</TabTrigger>
        <TabTrigger value="architecture">Architecture</TabTrigger>
        <TabTrigger value="flowchart">Flowchart</TabTrigger>
        <TabTrigger value="issues">Issues</TabTrigger>
        <TabTrigger value="recommendations">Recommendations</TabTrigger>
      </Tabs.List>

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
                  className={`panel-row${item.id === profile.id ? " border-zinc-800 bg-zinc-950" : ""}`}
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
            diagnostics={providerDiagnostics}
            onAPIKeyChange={onAPIKeyChange}
            onChange={onProfileChange}
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
          <section className="panel-block">
            <p className="panel-kicker">Flowchart</p>
            <h2 className="mt-3 text-2xl font-semibold tracking-tight text-zinc-50">
              {bundle ? `${bundle.summary.project_name || bundle.summary.name} graph` : "Project graph"}
            </h2>
            <p className="mt-3 text-sm leading-6 text-zinc-400">
              {bundle ? "Interactive topology rendered directly in the workbench." : "Run an analysis to unlock the project graph."}
            </p>
          </section>
          <FlowchartCanvas bundle={bundle} onInspect={onInspect} />
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

function TabTrigger({ children, value }: { children: ReactNode; value: TabKey }) {
  return (
    <Tabs.Trigger className="tab-trigger" value={value}>
      {children}
    </Tabs.Trigger>
  );
}
