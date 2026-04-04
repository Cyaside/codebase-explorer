import { useEffect } from "react";

import type { TabKey } from "@/lib/types";

interface UseWorkbenchShortcutsOptions {
  activeTab: TabKey;
  busy: boolean;
  onAnalyze: () => void;
  onCancel: () => void;
  onFocusAnalyze: () => void;
  onRefresh: () => void;
  onSelectTab: (tab: TabKey) => void;
  tabs: TabKey[];
}

export function useWorkbenchShortcuts({
  activeTab,
  busy,
  onAnalyze,
  onCancel,
  onFocusAnalyze,
  onRefresh,
  onSelectTab,
  tabs,
}: UseWorkbenchShortcutsOptions) {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const key = event.key;
      const hasCommandModifier = event.metaKey || event.ctrlKey;

      if (hasCommandModifier && key === "Enter") {
        event.preventDefault();
        onAnalyze();
        return;
      }

      if (key === "Escape" && busy) {
        event.preventDefault();
        onCancel();
        return;
      }

      if (isTypingTarget(event.target)) {
        return;
      }

      switch (key) {
        case "a":
        case "A":
        case "/":
          event.preventDefault();
          onFocusAnalyze();
          return;
        case "[":
          event.preventDefault();
          onSelectTab(stepTab(tabs, activeTab, -1));
          return;
        case "]":
          event.preventDefault();
          onSelectTab(stepTab(tabs, activeTab, 1));
          return;
        case "r":
        case "R":
          event.preventDefault();
          onRefresh();
          return;
        default:
          return;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activeTab, busy, onAnalyze, onCancel, onFocusAnalyze, onRefresh, onSelectTab, tabs]);
}

function isTypingTarget(target: EventTarget | null) {
  const element = target as HTMLElement | null;
  if (!element) {
    return false;
  }

  return (
    element instanceof HTMLInputElement ||
    element instanceof HTMLTextAreaElement ||
    element instanceof HTMLSelectElement ||
    element.isContentEditable
  );
}

function stepTab(tabs: TabKey[], activeTab: TabKey, direction: -1 | 1) {
  const index = tabs.indexOf(activeTab);
  if (index === -1 || tabs.length === 0) {
    return tabs[0] || activeTab;
  }

  const nextIndex = (index + direction + tabs.length) % tabs.length;
  return tabs[nextIndex];
}
