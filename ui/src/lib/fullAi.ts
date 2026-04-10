import type { FullAIResult, WorkbenchBundle } from "@/lib/types";

export function fullAIResult(bundle: WorkbenchBundle | null, name: string): FullAIResult | null {
  return bundle?.data.full_ai_execution.results.find((result) => result.name === name) || null;
}

export function hasFullAIOutput(result: FullAIResult | null): result is FullAIResult {
  if (!result || result.status !== "succeeded") {
    return false;
  }

  return Boolean(
    result.output.summary ||
      result.output.key_findings.length ||
      result.output.recommendations.length ||
      result.output.graph_edges.length ||
      result.output.issue_signals.length,
  );
}

export function fullAIStatusLabel(bundle: WorkbenchBundle | null): string {
  const summary = bundle?.data.full_ai;
  if (!summary?.enabled) {
    return "disabled";
  }

  return summary.status || bundle?.data.full_ai_execution.status || "prepared";
}

export function fullAINote(bundle: WorkbenchBundle | null): string {
  return bundle?.data.full_ai_execution.note || bundle?.data.full_ai.note || "";
}

