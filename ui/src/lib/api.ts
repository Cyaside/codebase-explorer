import type {
  AnalyzeResponse,
  AnalyzeRun,
  BundleSummary,
  ConnectionProfile,
  ProviderModelsResponse,
  ProviderTestResponse,
  WorkbenchBundle,
  WorkbenchStatusResponse,
} from "@/lib/types";
import {
  normalizeAnalyzeResponse,
  normalizeAnalyzeRun,
  normalizeBundleSummary,
  normalizeWorkbenchBundle,
  normalizeWorkbenchStatusResponse,
} from "@/lib/normalize";

export interface AnalyzePayload {
  repo_path: string;
  deterministic_only: boolean;
  ai_mode: "standard" | "full-ai";
  ai_read_budget: number;
  ai_token_budget: number;
  support_files: string[];
  extra_ignore_patterns: string[];
  provider: {
    name: string;
    model: string;
    api_key: string;
    base_url: string;
  } | null;
}

export interface ProviderDiagnosticPayload {
  provider: {
    name: string;
    model: string;
    api_key: string;
    base_url: string;
  };
}

async function requestJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = (await response.json().catch(() => ({}))) as Record<string, string>;
  if (!response.ok) {
    throw new Error(data.error || `Request failed with ${response.status}`);
  }
  return data as T;
}

export function fetchStatus() {
  return requestJSON<WorkbenchStatusResponse>("/api/status").then(normalizeWorkbenchStatusResponse);
}

export function fetchBundle(bundleName: string) {
  return requestJSON<WorkbenchBundle>(`/api/bundles/${encodeURIComponent(bundleName)}`).then(normalizeWorkbenchBundle);
}

export function deleteBundle(bundleName: string) {
  return requestJSON<{ deleted: string }>(`/api/bundles/${encodeURIComponent(bundleName)}`, {
    method: "DELETE",
  });
}

export function analyzeRepository(payload: AnalyzePayload) {
  return requestJSON<AnalyzeResponse>("/api/analyze", {
    method: "POST",
    body: JSON.stringify(payload),
  }).then(normalizeAnalyzeResponse);
}

export function startAnalyzeRun(payload: AnalyzePayload) {
  return requestJSON<{ run: AnalyzeRun }>("/api/analyze-runs", {
    method: "POST",
    body: JSON.stringify(payload),
  }).then((response) => ({
    run: normalizeAnalyzeRun(response.run),
  }));
}

export function fetchAnalyzeRun(runID: string) {
  return requestJSON<{ run: AnalyzeRun }>(`/api/analyze-runs/${encodeURIComponent(runID)}`).then((response) => ({
    run: normalizeAnalyzeRun(response.run),
  }));
}

export function cancelAnalyzeRun(runID: string) {
  return requestJSON<{ run: AnalyzeRun }>(`/api/analyze-runs/${encodeURIComponent(runID)}/cancel`, {
    method: "POST",
  }).then((response) => ({
    run: normalizeAnalyzeRun(response.run),
  }));
}

export function fetchProviderModels(payload: ProviderDiagnosticPayload) {
  return requestJSON<ProviderModelsResponse>("/api/provider/models", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function testProvider(payload: ProviderDiagnosticPayload) {
  return requestJSON<ProviderTestResponse>("/api/provider/test", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function buildProviderPayload(profile: ConnectionProfile, apiKey: string) {
  return {
    name: profile.provider.name,
    model: profile.provider.model,
    api_key: apiKey.trim(),
    base_url: profile.provider.baseUrl,
  };
}

export function mergeBundleSummary(existing: BundleSummary[], next: BundleSummary) {
  const normalized = normalizeBundleSummary(next);
  return [normalized, ...existing.filter((bundle) => bundle.name !== normalized.name)];
}
