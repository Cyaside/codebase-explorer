import type {
  AnalyzeResponse,
  AnalyzeRun,
  BundleSummary,
  ConnectionProfile,
  WorkbenchBundle,
  WorkbenchStatusResponse,
} from "@/lib/types";

export interface AnalyzePayload {
  repo_path: string;
  deterministic_only: boolean;
  support_files: string[];
  extra_ignore_patterns: string[];
  provider: {
    name: string;
    model: string;
    api_key: string;
    base_url: string;
  } | null;
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
  return requestJSON<WorkbenchStatusResponse>("/api/status");
}

export function fetchBundle(bundleName: string) {
  return requestJSON<WorkbenchBundle>(`/api/bundles/${encodeURIComponent(bundleName)}`);
}

export function analyzeRepository(payload: AnalyzePayload) {
  return requestJSON<AnalyzeResponse>("/api/analyze", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function startAnalyzeRun(payload: AnalyzePayload) {
  return requestJSON<{ run: AnalyzeRun }>("/api/analyze-runs", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function fetchAnalyzeRun(runID: string) {
  return requestJSON<{ run: AnalyzeRun }>(`/api/analyze-runs/${encodeURIComponent(runID)}`);
}

export function cancelAnalyzeRun(runID: string) {
  return requestJSON<{ run: AnalyzeRun }>(`/api/analyze-runs/${encodeURIComponent(runID)}/cancel`, {
    method: "POST",
  });
}

export function buildProviderPayload(profile: ConnectionProfile, apiKey: string) {
  if (!profile.provider) {
    return null;
  }

  return {
    name: profile.provider.name,
    model: profile.provider.model,
    api_key: apiKey.trim(),
    base_url: profile.provider.baseUrl,
  };
}

export function mergeBundleSummary(existing: BundleSummary[], next: BundleSummary) {
  return [next, ...existing.filter((bundle) => bundle.name !== next.name)];
}
