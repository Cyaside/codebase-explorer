import { useEffect } from "react";

import type { TabKey } from "@/lib/types";

interface UseWorkbenchShortcutsOptions {
  activeTab: TabKey;
  busy: boolean;
  commandPaletteOpen: boolean;
  onNextBundle: () => void;
  onPreviousBundle: () => void;
  onAnalyze: () => void;
  onCancel: () => void;
  onCloseCommandPalette: () => void;
  onFocusAnalyze: () => void;
  onOpenCommandPalette: () => void;
  onRefresh: () => void;
  onSelectTab: (tab: TabKey) => void;
  tabs: TabKey[];
}

export function useWorkbenchShortcuts({
  activeTab,
  busy,
  commandPaletteOpen,
  onNextBundle,
  onPreviousBundle,
  onAnalyze,
  onCancel,
  onCloseCommandPalette,
  onFocusAnalyze,
  onOpenCommandPalette,
  onRefresh,
  onSelectTab,
  tabs,
}: UseWorkbenchShortcutsOptions) {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const key = event.key;
      const hasCommandModifier = event.metaKey || event.ctrlKey;
      const keyLower = key.toLowerCase();

      if (hasCommandModifier && keyLower === "k") {
        event.preventDefault();
        onOpenCommandPalette();
        return;
      }

      if (commandPaletteOpen) {
        if (key === "Escape") {
          event.preventDefault();
          onCloseCommandPalette();
        }
        return;
      }

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
        case "j":
        case "J":
          event.preventDefault();
          onNextBundle();
          return;
        case "k":
        case "K":
          event.preventDefault();
          onPreviousBundle();
          return;
        default:
          return;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    activeTab,
    busy,
    commandPaletteOpen,
    onAnalyze,
    onCancel,
    onCloseCommandPalette,
    onFocusAnalyze,
    onNextBundle,
    onOpenCommandPalette,
    onPreviousBundle,
    onRefresh,
    onSelectTab,
    tabs,
  ]);
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
