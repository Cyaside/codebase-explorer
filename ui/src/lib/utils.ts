export function cn(...inputs: Array<string | false | null | undefined>) {
  return inputs.filter(Boolean).join(" ");
}

export function bundleLink(bundleName: string, relativePath: string) {
  return `/bundles/${encodeURIComponent(bundleName)}/${relativePath}`;
}

export function normalizeLocalPath(value: string) {
  return value.trim().replace(/\//g, "\\").toLowerCase();
}

export function matchesWorkspace(bundlePath: string, workspacePath: string) {
  return !!workspacePath.trim() && normalizeLocalPath(bundlePath) === normalizeLocalPath(workspacePath);
}

export function formatTimestamp(value: string) {
  if (!value) {
    return "Unknown";
  }

  const timestamp = new Date(value);
  if (Number.isNaN(timestamp.getTime())) {
    return value;
  }

  return timestamp.toLocaleString();
}

export function formatRelativeTime(value: string) {
  if (!value) {
    return "just now";
  }

  const timestamp = new Date(value);
  if (Number.isNaN(timestamp.getTime())) {
    return value;
  }

  const diff = Date.now() - timestamp.getTime();
  const minutes = Math.round(diff / 60000);
  if (minutes <= 1) {
    return "just now";
  }
  if (minutes < 60) {
    return `${minutes}m ago`;
  }

  const hours = Math.round(minutes / 60);
  if (hours < 24) {
    return `${hours}h ago`;
  }

  const days = Math.round(hours / 24);
  return `${days}d ago`;
}
