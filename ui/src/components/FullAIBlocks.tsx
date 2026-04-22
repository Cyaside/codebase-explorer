import { CheckCircle2, CircleAlert, ShieldCheck } from "lucide-react";

import { fullAINote, fullAIStatusLabel, hasFullAIOutput } from "@/lib/fullAi";
import type { FullAIResult, InspectorState, WorkbenchBundle } from "@/lib/types";

interface FullAITextPanelProps {
  bundle: WorkbenchBundle;
  eyebrow: string;
  fallbackBody: string;
  result: FullAIResult | null;
  title: string;
}

export function FullAITextPanel({ bundle, eyebrow, fallbackBody, result, title }: FullAITextPanelProps) {
  const ready = hasFullAIOutput(result);
  const body = ready ? result.output.summary || fallbackBody : fallbackBody;
  const note = ready ? `${result.name} verified: ${result.verified ? "yes" : "needs review"}` : fullAINote(bundle);

  return (
    <section className="panel-block">
      <div className="flex items-center justify-between gap-3">
        <p className="panel-kicker">{eyebrow}</p>
        <StatusPill status={ready ? result.status : fullAIStatusLabel(bundle)} />
      </div>
      <h3 className="mt-3 text-lg font-semibold text-zinc-100">{title}</h3>
      <p className="mt-4 text-sm leading-7 text-zinc-300">{body || "No narrative available yet."}</p>
      {note ? <p className="mt-4 text-xs leading-5 text-zinc-500">{note}</p> : null}
      {ready && result.output.key_findings.length ? (
        <div className="mt-5 space-y-2 border-t border-zinc-900 pt-4">
          {result.output.key_findings.slice(0, 5).map((finding) => (
            <div className="rounded-2xl border border-zinc-900 bg-black px-3 py-3" key={finding.claim}>
              <p className="text-sm leading-6 text-zinc-200">{finding.claim}</p>
              <p className="mt-2 text-xs text-zinc-600">{finding.confidence || "confidence not specified"}</p>
            </div>
          ))}
        </div>
      ) : null}
    </section>
  );
}

export function FullAIStatusPanel({ bundle }: { bundle: WorkbenchBundle }) {
  const summary = bundle.data.full_ai;
  const execution = bundle.data.full_ai_execution;
  const verification = bundle.data.full_ai_verification;
  const reviewFunctions = verification.functions.filter((item) => !item.verified || item.warnings.length).slice(0, 5);
  const stats = [
    { label: "Status", value: fullAIStatusLabel(bundle) },
    { label: "Evidence read", value: `${summary.collected_items}/${summary.planned_targets}` },
    { label: "Functions", value: `${execution.executed_count}/${summary.prepared_functions}` },
    { label: "Verified", value: `${verification.verified_count || execution.verified_count}` },
    { label: "Failed checks", value: `${verification.failed_check_count}` },
    { label: "Warnings", value: `${verification.warning_count}` },
  ];

  return (
    <section className="panel-block">
      <div className="flex items-center gap-3">
        <ShieldCheck className="size-4 text-zinc-500" />
        <p className="panel-kicker">Full-AI runtime</p>
      </div>
      <div className="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-6">
        {stats.map((item) => (
          <div className="rounded-2xl border border-zinc-900 bg-black px-3 py-3" key={item.label}>
            <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-zinc-600">{item.label}</p>
            <p className="mt-2 text-sm font-semibold text-zinc-100">{item.value}</p>
          </div>
        ))}
      </div>
      {reviewFunctions.length ? (
        <div className="mt-4 space-y-1.5 border-t border-zinc-900 pt-4">
          {reviewFunctions.map((item) => {
            const failedCheck = item.checks.find((check) => check.status === "fail");
            const warning = item.warnings[0];
            return (
              <div className="panel-row" key={item.name}>
                <div className="min-w-0">
                  <p className="text-sm font-medium text-zinc-100">{item.name}</p>
                  <p className="mt-1 line-clamp-2 text-xs leading-5 text-zinc-500">
                    {failedCheck?.detail || warning || "Verification needs review."}
                  </p>
                </div>
                <span className="shrink-0 text-xs text-zinc-500">{item.status || "review"}</span>
              </div>
            );
          })}
        </div>
      ) : null}
      {execution.note || summary.note ? <p className="mt-4 text-xs leading-5 text-zinc-500">{execution.note || summary.note}</p> : null}
      {verification.note ? <p className="mt-2 text-xs leading-5 text-zinc-600">{verification.note}</p> : null}
    </section>
  );
}

export function FullAIIssueSignalsPanel({
  onInspect,
  result,
}: {
  onInspect: (value: InspectorState | null) => void;
  result: FullAIResult | null;
}) {
  if (!hasFullAIOutput(result) || !result.output.issue_signals.length) {
    return null;
  }

  return (
    <section className="panel-block">
      <div className="flex items-center gap-3">
        <CircleAlert className="size-4 text-zinc-500" />
        <p className="panel-kicker">AI issue signals</p>
      </div>
      <div className="mt-4 space-y-1.5">
        {result.output.issue_signals.map((signal) => (
          <button
            className="panel-row"
            key={`${signal.title}-${signal.severity}`}
            onClick={() =>
              onInspect({
                eyebrow: "AI issue signal",
                title: signal.title,
                description: signal.severity || "Severity not specified.",
                notes: signal.evidence_paths.length ? signal.evidence_paths : ["No evidence paths attached."],
                properties: [{ label: "Function", value: result.name }],
              })
            }
            type="button"
          >
            <div className="min-w-0">
              <p className="truncate text-sm font-medium text-zinc-100">{signal.title}</p>
              <p className="mt-1 text-xs text-zinc-500">{signal.evidence_paths.slice(0, 2).join(", ") || "No evidence paths"}</p>
            </div>
            <span className="shrink-0 text-xs text-zinc-500">{signal.severity || "signal"}</span>
          </button>
        ))}
      </div>
    </section>
  );
}

export function FullAIRecommendationsPanel({
  onInspect,
  result,
}: {
  onInspect: (value: InspectorState | null) => void;
  result: FullAIResult | null;
}) {
  if (!hasFullAIOutput(result) || !result.output.recommendations.length) {
    return null;
  }

  return (
    <section className="panel-block">
      <div className="flex items-center gap-3">
        <CheckCircle2 className="size-4 text-zinc-500" />
        <p className="panel-kicker">AI recommendations</p>
      </div>
      <div className="mt-4 space-y-1.5">
        {result.output.recommendations.map((recommendation, index) => (
          <button
            className="panel-row"
            key={`${recommendation}-${index}`}
            onClick={() =>
              onInspect({
                eyebrow: "AI recommendation",
                title: `Recommendation ${index + 1}`,
                description: recommendation,
                notes: result.output.uncertainties,
                properties: [{ label: "Function", value: result.name }],
              })
            }
            type="button"
          >
            <span className="text-sm leading-6 text-zinc-200">{recommendation}</span>
            <span className="shrink-0 text-xs text-zinc-500">AI</span>
          </button>
        ))}
      </div>
    </section>
  );
}

function StatusPill({ status }: { status: string }) {
  return <span className="run-pill">{status || "unknown"}</span>;
}
