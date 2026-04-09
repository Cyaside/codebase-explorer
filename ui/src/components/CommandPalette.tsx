import { useEffect, useRef, useState } from "react";
import { Command, CornerDownLeft, Search } from "lucide-react";

export interface CommandPaletteAction {
  id: string;
  group: string;
  label: string;
  description: string;
  keywords?: string[];
  shortcut?: string;
  run: () => void;
}

interface CommandPaletteProps {
  actions: CommandPaletteAction[];
  open: boolean;
  onClose: () => void;
}

export function CommandPalette({ actions, open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (!open) {
      setQuery("");
      setActiveIndex(0);
      return;
    }

    const timeoutID = window.setTimeout(() => {
      inputRef.current?.focus();
      inputRef.current?.select();
    }, 10);

    return () => window.clearTimeout(timeoutID);
  }, [open]);

  const normalizedQuery = query.trim().toLowerCase();
  const filtered = !normalizedQuery
    ? actions
    : actions.filter((action) => {
        const haystack = [action.group, action.label, action.description, ...(action.keywords || [])].join(" ").toLowerCase();
        return haystack.includes(normalizedQuery);
      });

  useEffect(() => {
    if (activeIndex >= filtered.length) {
      setActiveIndex(filtered.length ? filtered.length - 1 : 0);
    }
  }, [activeIndex, filtered.length]);

  if (!open) {
    return null;
  }

  const grouped = groupActions(filtered);

  const handleAction = (action: CommandPaletteAction) => {
    action.run();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-[70] bg-black/72 backdrop-blur-sm" onClick={onClose} role="presentation">
      <div className="mx-auto mt-[10vh] w-full max-w-3xl px-4" onClick={(event) => event.stopPropagation()} role="presentation">
        <section className="rounded-[1.75rem] border border-zinc-900 bg-zinc-950 shadow-2xl">
          <div className="flex items-center gap-3 border-b border-zinc-900 px-4 py-4">
            <Command className="size-4 text-zinc-400" />
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-zinc-600" />
              <input
                className="compact-input pl-10"
                onChange={(event) => {
                  setQuery(event.target.value);
                  setActiveIndex(0);
                }}
                onKeyDown={(event) => {
                  if (event.key === "ArrowDown") {
                    event.preventDefault();
                    setActiveIndex((current) => (filtered.length ? (current + 1) % filtered.length : 0));
                    return;
                  }
                  if (event.key === "ArrowUp") {
                    event.preventDefault();
                    setActiveIndex((current) => (filtered.length ? (current - 1 + filtered.length) % filtered.length : 0));
                    return;
                  }
                  if (event.key === "Enter") {
                    event.preventDefault();
                    const action = filtered[activeIndex];
                    if (action) {
                      handleAction(action);
                    }
                    return;
                  }
                  if (event.key === "Escape") {
                    event.preventDefault();
                    onClose();
                  }
                }}
                placeholder="Search projects, bundles, and commands"
                ref={inputRef}
                type="text"
                value={query}
              />
            </div>
            <span className="hidden rounded-full border border-zinc-800 bg-black px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500 md:inline-flex">
              Cmd/Ctrl+K
            </span>
          </div>

          <div className="max-h-[65vh] overflow-y-auto px-2 py-2">
            {filtered.length ? (
              grouped.map(([group, groupActions]) => (
                <div className="px-2 py-2" key={group}>
                  <p className="px-2 pb-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-zinc-600">{group}</p>
                  <div className="space-y-1">
                    {groupActions.map((action) => {
                      const globalIndex = filtered.findIndex((item) => item.id === action.id);
                      const selected = globalIndex === activeIndex;
                      return (
                        <button className={selected ? "command-item command-item-active" : "command-item"} key={action.id} onClick={() => handleAction(action)} type="button">
                          <div className="min-w-0">
                            <p className="truncate text-sm font-semibold text-zinc-100">{action.label}</p>
                            <p className="mt-1 truncate text-sm text-zinc-500">{action.description}</p>
                          </div>
                          <div className="ml-3 flex shrink-0 items-center gap-2">
                            {action.shortcut ? <span className="hint-chip">{action.shortcut}</span> : null}
                            <CornerDownLeft className="size-4 text-zinc-600" />
                          </div>
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))
            ) : (
              <div className="px-4 py-10 text-center">
                <p className="text-sm font-semibold text-zinc-100">No matching command</p>
                <p className="mt-2 text-sm text-zinc-500">Try a project name, a bundle name, or an action like "refresh" or "flowchart".</p>
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
}

function groupActions(actions: CommandPaletteAction[]) {
  const grouped = new Map<string, CommandPaletteAction[]>();
  for (const action of actions) {
    const bucket = grouped.get(action.group) || [];
    bucket.push(action);
    grouped.set(action.group, bucket);
  }
  return Array.from(grouped.entries());
}
