import type { AnalyzeResponse, AnalyzeRun, BundleData, BundleSummary, WorkbenchBundle, WorkbenchStatusResponse } from "@/lib/types";

type UnknownRecord = Record<string, unknown>;

export function normalizeWorkbenchStatusResponse(value: unknown): WorkbenchStatusResponse {
  const record = asRecord(value);
  return {
    app_version: asString(record.app_version),
    output_root: asString(record.output_root),
    cache_root: asString(record.cache_root),
    cache_enabled: asBoolean(record.cache_enabled),
    default_provider: {
      name: asString(asRecord(record.default_provider).name),
      model: asString(asRecord(record.default_provider).model),
      base_url: asString(asRecord(record.default_provider).base_url),
    },
    supported_providers: asArray(record.supported_providers).map((item) => {
      const provider = asRecord(item);
      return {
        name: asString(provider.name),
        requires_api_key: asBoolean(provider.requires_api_key),
        requires_model: asBoolean(provider.requires_model),
        requires_base_url: asBoolean(provider.requires_base_url),
      };
    }),
    recent_bundles: asArray(record.recent_bundles).map(normalizeBundleSummary),
    bundle_warnings: asStringArray(record.bundle_warnings),
  };
}

export function normalizeWorkbenchBundle(value: unknown): WorkbenchBundle {
  const record = asRecord(value);
  return {
    summary: normalizeBundleSummary(record.summary),
    data: normalizeBundleData(record.data),
  };
}

export function normalizeAnalyzeResponse(value: unknown): AnalyzeResponse {
  const record = asRecord(value);
  const result = asRecord(record.result);

  return {
    result: {
      output_path: asString(result.output_path),
      project_name: asString(result.project_name),
      project_type: asString(result.project_type),
      total_files: asNumber(result.total_files),
      total_lines: asNumber(result.total_lines),
      primary_language: asString(result.primary_language),
    },
    bundle: normalizeBundleSummary(record.bundle),
    data: normalizeBundleData(record.data),
  };
}

export function normalizeAnalyzeRun(value: unknown): AnalyzeRun {
  const record = asRecord(value);
  return {
    id: asString(record.id),
    status: asString(record.status),
    created_at: asString(record.created_at),
    updated_at: asString(record.updated_at),
    progress: asArray(record.progress).map((item) => {
      const event = asRecord(item);
      return {
        stage: asString(event.stage),
        status: asString(event.status),
        detail: asString(event.detail),
      };
    }),
    error: asOptionalString(record.error),
    response: record.response ? normalizeAnalyzeResponse(record.response) : undefined,
  };
}

export function normalizeBundleSummary(value: unknown): BundleSummary {
  const record = asRecord(value);
  return {
    name: asString(record.name),
    path: asString(record.path),
    generated_at: asString(record.generated_at),
    project_name: asString(record.project_name),
    project_type: asString(record.project_type),
    analyzed_path: asString(record.analyzed_path),
    total_files: asNumber(record.total_files),
    total_lines: asNumber(record.total_lines),
    ai_status: asString(record.ai_status),
    warning_count: asNumber(record.warning_count),
    support_file_count: asNumber(record.support_file_count),
  };
}

export function normalizeBundleData(value: unknown): BundleData {
  const record = asRecord(value);
  const project = asRecord(record.project);
  const metrics = asRecord(record.metrics);
  const changes = asRecord(record.changes);
  const ai = asRecord(record.ai);
  const fullAI = asRecord(record.full_ai);
  const fullAIExecution = asRecord(record.full_ai_execution);
  const mermaid = asRecord(record.mermaid);
  const links = asRecord(record.links);

  const projectSummary = asString(project.summary);
  const fallbackReadingPath = asArray(record.reading_path).map((item) => {
    const readingItem = asRecord(item);
    return {
      path: asString(readingItem.path),
      reason: asString(readingItem.reason),
    };
  });

  return {
    bundle_name: asString(record.bundle_name),
    generated_at: asString(record.generated_at),
    project: {
      name: asString(project.name),
      type: asString(project.type),
      analyzed_path: asString(project.analyzed_path),
      summary: projectSummary,
      provider_mode: asString(project.provider_mode),
    },
    metrics: {
      total_files: asNumber(metrics.total_files),
      total_lines: asNumber(metrics.total_lines),
    },
    warnings: asStringArray(record.warnings),
    languages: asArray(record.languages).map((item) => {
      const language = asRecord(item);
      return {
        name: asString(language.name),
        file_count: asNumber(language.file_count),
        line_count: asNumber(language.line_count),
      };
    }),
    important_directories: asStringArray(record.important_directories),
    entry_points: asStringArray(record.entry_points),
    core_modules: asStringArray(record.core_modules),
    modules: asArray(record.modules).map((item) => {
      const module = asRecord(item);
      return {
        path: asString(module.path),
        file_count: asNumber(module.file_count),
        total_lines: asNumber(module.total_lines),
        entry_point_count: asNumber(module.entry_point_count),
      };
    }),
    reading_path: fallbackReadingPath,
    changes: {
      note: asString(changes.note),
      support_file_count: asNumber(changes.support_file_count),
      sources: asArray(changes.sources).map((item) => {
        const source = asRecord(item);
        return {
          path: asString(source.path),
          kind: asString(source.kind),
          status: asString(source.status),
        };
      }),
      frequently_mentioned_areas: asArray(changes.frequently_mentioned_areas).map((item) => {
        const area = asRecord(item);
        return {
          path: asString(area.path),
          mention_count: asNumber(area.mention_count),
          confidence: asString(area.confidence),
        };
      }),
      repeated_themes: asArray(changes.repeated_themes).map((item) => {
        const theme = asRecord(item);
        return {
          name: asString(theme.name),
          mention_count: asNumber(theme.mention_count),
        };
      }),
    },
    ai: {
      status: asString(ai.status),
      provider: asString(ai.provider),
      model: asString(ai.model),
      note: asString(ai.note || ai.fallback_reason),
      project_summary: asString(ai.project_summary || projectSummary),
      architecture_narrative: asString(ai.architecture_narrative),
      hotspot_explanations: asArray(ai.hotspot_explanations).map((item) => {
        const hotspot = asRecord(item);
        return {
          path: asString(hotspot.path),
          explanation: asString(hotspot.explanation),
        };
      }),
      reading_path_explanations: asArray(ai.reading_path_explanations).map((item) => {
        const readingItem = asRecord(item);
        return {
          path: asString(readingItem.path),
          rationale: asString(readingItem.rationale),
        };
      }),
    },
    full_ai: normalizeFullAISummary(fullAI),
    full_ai_execution: normalizeFullAIExecution(fullAIExecution),
    mermaid: {
      architecture: asString(mermaid.architecture),
      dependencies: asString(mermaid.dependencies),
    },
    links: {
      architecture_diagram: asString(links.architecture_diagram),
      dependency_diagram: asString(links.dependency_diagram),
    },
  };
}

function normalizeFullAISummary(record: UnknownRecord) {
  return {
    enabled: asBoolean(record.enabled),
    mode: asString(record.mode),
    status: asString(record.status),
    read_budget: asNumber(record.read_budget),
    token_budget: asNumber(record.token_budget),
    planned_targets: asNumber(record.planned_targets),
    planned_functions: asNumber(record.planned_functions),
    collected_items: asNumber(record.collected_items),
    failed_items: asNumber(record.failed_items),
    prepared_functions: asNumber(record.prepared_functions),
    executed_functions: asNumber(record.executed_functions),
    verified_functions: asNumber(record.verified_functions),
    note: asString(record.note),
  };
}

function normalizeFullAIExecution(record: UnknownRecord) {
  return {
    mode: asString(record.mode),
    provider: asString(record.provider),
    model: asString(record.model),
    status: asString(record.status),
    executed_count: asNumber(record.executed_count),
    verified_count: asNumber(record.verified_count),
    failed_count: asNumber(record.failed_count),
    note: asString(record.note),
    results: asArray(record.results).map((item) => {
      const result = asRecord(item);
      const output = asRecord(result.output);
      return {
        name: asString(result.name),
        status: asString(result.status),
        instruction_path: asString(result.instruction_path),
        evidence_paths: asStringArray(result.evidence_paths),
        output: {
          summary: asString(output.summary),
          key_findings: asArray(output.key_findings).map((findingItem) => {
            const finding = asRecord(findingItem);
            return {
              claim: asString(finding.claim),
              evidence_paths: asStringArray(finding.evidence_paths),
              confidence: asString(finding.confidence),
            };
          }),
          recommendations: asStringArray(output.recommendations),
          graph_edges: asArray(output.graph_edges).map((edgeItem) => {
            const edge = asRecord(edgeItem);
            return {
              from: asString(edge.from),
              to: asString(edge.to),
              label: asString(edge.label),
              evidence_paths: asStringArray(edge.evidence_paths),
            };
          }),
          issue_signals: asArray(output.issue_signals).map((signalItem) => {
            const signal = asRecord(signalItem);
            return {
              title: asString(signal.title),
              severity: asString(signal.severity),
              evidence_paths: asStringArray(signal.evidence_paths),
            };
          }),
          uncertainties: asStringArray(output.uncertainties),
        },
        verified: asBoolean(result.verified),
        error: asString(result.error),
      };
    }),
  };
}

function asRecord(value: unknown): UnknownRecord {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return value as UnknownRecord;
}

function asArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function asOptionalString(value: unknown): string | undefined {
  return typeof value === "string" ? value : undefined;
}

function asStringArray(value: unknown): string[] {
  return asArray(value).filter((item): item is string => typeof item === "string");
}

function asNumber(value: unknown): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function asBoolean(value: unknown): boolean {
  return value === true;
}
