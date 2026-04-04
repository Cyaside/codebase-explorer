import type { ReactNode } from "react";
import { AlertCircle, CircleAlert, OctagonAlert, TriangleAlert } from "lucide-react";

import type { AnalyzeRun, WorkbenchBundle, WorkbenchStatusResponse } from "@/lib/types";

interface WorkbenchAlertsProps {
  activeRun: AnalyzeRun | null;
  bundle: WorkbenchBundle | null;
  errorMessage: string;
  status: WorkbenchStatusResponse | null;
}

export function WorkbenchAlerts({ activeRun, bundle, errorMessage, status }: WorkbenchAlertsProps) {
  const alerts = buildAlerts(activeRun, bundle, errorMessage, status);
  if (!alerts.length) {
    return null;
  }

  return (
    <section className="space-y-3">
      {alerts.map((alert) => (
        <article className={alert.panelClassName} key={`${alert.level}-${alert.title}`}>
          <div className="flex items-start gap-3">
            <span className={alert.iconClassName}>{alert.icon}</span>
            <div className="min-w-0">
              <p className={alert.titleClassName}>{alert.title}</p>
              <p className={alert.bodyClassName}>{alert.body}</p>
              {alert.items.length ? (
                <ul className="mt-3 space-y-2">
                  {alert.items.map((item) => (
                    <li className={alert.itemClassName} key={item}>
                      {item}
                    </li>
                  ))}
                </ul>
              ) : null}
            </div>
          </div>
        </article>
      ))}
    </section>
  );
}

function buildAlerts(
  activeRun: AnalyzeRun | null,
  bundle: WorkbenchBundle | null,
  errorMessage: string,
  status: WorkbenchStatusResponse | null,
) {
  const alerts: AlertRecord[] = [];

  if (errorMessage) {
    alerts.push({
      level: "error",
      title: "Workbench needs attention",
      body: errorMessage,
      icon: <AlertCircle className="size-4" />,
      items: [],
    });
  }

  if (activeRun?.status === "canceling") {
    alerts.push({
      level: "info",
      title: "Cancel requested",
      body: "The active analyze run is shutting down. Current work will stop as soon as the backend reaches a safe cancellation point.",
      icon: <CircleAlert className="size-4" />,
      items: [],
    });
  }

  if (status?.bundle_warnings?.length) {
    alerts.push({
      level: "warning",
      title: "Bundle inventory has skipped entries",
      body: "One or more local bundles could not be loaded cleanly. Review these before trusting the recent bundle list.",
      icon: <TriangleAlert className="size-4" />,
      items: status.bundle_warnings.slice(0, 4),
    });
  }

  if ((bundle?.data.warnings || []).length) {
    alerts.push({
      level: "warning",
      title: "Selected bundle contains runtime warnings",
      body: "This analysis completed, but parts of the bundle need extra scrutiny before you rely on the output.",
      icon: <OctagonAlert className="size-4" />,
      items: (bundle?.data.warnings || []).slice(0, 4),
    });
  }

  return alerts.map(applyAlertTheme);
}

interface AlertRecord {
  level: "error" | "warning" | "info";
  title: string;
  body: string;
  icon: ReactNode;
  items: string[];
}

function applyAlertTheme(alert: AlertRecord) {
  switch (alert.level) {
    case "error":
      return {
        ...alert,
        panelClassName: "rounded-2xl border border-danger/40 bg-danger/10 px-4 py-3",
        iconClassName: "mt-0.5 shrink-0 text-danger",
        titleClassName: "text-sm font-semibold text-danger",
        bodyClassName: "mt-1 text-sm text-danger/90",
        itemClassName: "rounded-xl border border-danger/20 bg-background/30 px-3 py-2 text-sm text-danger/85",
      };
    case "warning":
      return {
        ...alert,
        panelClassName: "rounded-2xl border border-amber-400/35 bg-amber-400/10 px-4 py-3",
        iconClassName: "mt-0.5 shrink-0 text-amber-300",
        titleClassName: "text-sm font-semibold text-amber-100",
        bodyClassName: "mt-1 text-sm text-amber-50/85",
        itemClassName: "rounded-xl border border-amber-300/20 bg-background/30 px-3 py-2 text-sm text-amber-50/80",
      };
    default:
      return {
        ...alert,
        panelClassName: "rounded-2xl border border-primary/30 bg-primary/10 px-4 py-3",
        iconClassName: "mt-0.5 shrink-0 text-primary",
        titleClassName: "text-sm font-semibold text-foreground",
        bodyClassName: "mt-1 text-sm text-muted-foreground",
        itemClassName: "rounded-xl border border-primary/20 bg-background/30 px-3 py-2 text-sm text-foreground/85",
      };
  }
}
