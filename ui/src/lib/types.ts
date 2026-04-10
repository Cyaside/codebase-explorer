export type TabKey =
  | "projects"
  | "project"
  | "connections"
  | "properties"
  | "dashboard"
  | "summary"
  | "architecture"
  | "flowchart"
  | "issues"
  | "recommendations";

export interface ProviderProfile {
  name: string;
  model: string;
  baseUrl: string;
}

export interface ConnectionProfile {
  id: string;
  label: string;
  provider: ProviderProfile;
  locked?: boolean;
}

export interface SavedWorkspace {
  id: string;
  label: string;
  repoPath: string;
  supportFiles: string[];
  ignorePatterns: string[];
  selectedProfile: string;
  activeBundle: string;
  createdAt: string;
  updatedAt: string;
}

export interface SupportedProviderOption {
  name: string;
  requires_api_key: boolean;
  requires_model: boolean;
  requires_base_url: boolean;
}

export interface WorkbenchStatusResponse {
  app_version: string;
  output_root: string;
  cache_root: string;
  cache_enabled: boolean;
  default_provider: {
    name: string;
    model: string;
    base_url: string;
  };
  supported_providers: SupportedProviderOption[];
  recent_bundles: BundleSummary[];
  bundle_warnings: string[];
}

export interface AnalyzeResponse {
  result: {
    output_path: string;
    project_name: string;
    project_type: string;
    total_files: number;
    total_lines: number;
    primary_language: string;
  };
  bundle: BundleSummary;
  data: BundleData;
}

export interface AnalyzeProgressEvent {
  stage: string;
  status: string;
  detail: string;
}

export interface AnalyzeRun {
  id: string;
  status: "queued" | "running" | "canceling" | "succeeded" | "failed" | "canceled" | string;
  created_at: string;
  updated_at: string;
  progress: AnalyzeProgressEvent[];
  error?: string;
  response?: AnalyzeResponse;
}

export interface BundleSummary {
  name: string;
  path: string;
  generated_at: string;
  project_name: string;
  project_type: string;
  analyzed_path: string;
  total_files: number;
  total_lines: number;
  ai_status: string;
  warning_count: number;
  support_file_count: number;
}

export interface WorkbenchBundle {
  summary: BundleSummary;
  data: BundleData;
}

export interface BundleData {
  bundle_name: string;
  generated_at: string;
  project: {
    name: string;
    type: string;
    analyzed_path: string;
    summary: string;
    provider_mode: string;
  };
  metrics: {
    total_files: number;
    total_lines: number;
  };
  warnings: string[];
  languages: Array<{
    name: string;
    file_count: number;
    line_count: number;
  }>;
  important_directories: string[];
  entry_points: string[];
  core_modules: string[];
  modules: Array<{
    path: string;
    file_count: number;
    total_lines: number;
    entry_point_count: number;
  }>;
  reading_path: Array<{
    path: string;
    reason: string;
  }>;
  changes: {
    note: string;
    support_file_count: number;
    sources: Array<{
      path: string;
      kind: string;
      status: string;
    }>;
    frequently_mentioned_areas: Array<{
      path: string;
      mention_count: number;
      confidence: string;
    }>;
    repeated_themes: Array<{
      name: string;
      mention_count: number;
    }>;
  };
  ai: {
    status: string;
    provider: string;
    model: string;
    note: string;
    project_summary: string;
    architecture_narrative: string;
    hotspot_explanations: Array<{
      path: string;
      explanation: string;
    }>;
    reading_path_explanations: Array<{
      path: string;
      rationale: string;
    }>;
  };
  full_ai: FullAISummary;
  full_ai_execution: FullAIExecution;
  mermaid: {
    architecture: string;
    dependencies: string;
  };
  links: {
    architecture_diagram: string;
    dependency_diagram: string;
  };
}

export interface FullAISummary {
  enabled: boolean;
  mode: string;
  status: string;
  read_budget: number;
  token_budget: number;
  planned_targets: number;
  planned_functions: number;
  collected_items: number;
  failed_items: number;
  prepared_functions: number;
  executed_functions: number;
  verified_functions: number;
  note: string;
}

export interface FullAIExecution {
  mode: string;
  provider: string;
  model: string;
  status: string;
  executed_count: number;
  verified_count: number;
  failed_count: number;
  note: string;
  results: FullAIResult[];
}

export interface FullAIResult {
  name: string;
  status: string;
  instruction_path: string;
  evidence_paths: string[];
  output: FullAIOutput;
  verified: boolean;
  error: string;
}

export interface FullAIOutput {
  summary: string;
  key_findings: FullAIFinding[];
  recommendations: string[];
  graph_edges: FullAIGraphEdge[];
  issue_signals: FullAIIssueSignal[];
  uncertainties: string[];
}

export interface FullAIFinding {
  claim: string;
  evidence_paths: string[];
  confidence: string;
}

export interface FullAIGraphEdge {
  from: string;
  to: string;
  label: string;
  evidence_paths: string[];
}

export interface FullAIIssueSignal {
  title: string;
  severity: string;
  evidence_paths: string[];
}

export interface PersistedUIState {
  activeTab: TabKey;
  activeWorkspace: string;
}

export interface InspectorProperty {
  label: string;
  value: string;
}

export interface InspectorState {
  eyebrow: string;
  title: string;
  description: string;
  properties: InspectorProperty[];
  notes: string[];
}
