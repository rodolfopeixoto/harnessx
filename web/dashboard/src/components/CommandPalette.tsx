import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { tokens } from "../ds/tokens";

type Hit = {
  source: string;
  kind: string;
  title: string;
  subtitle?: string;
  router_path?: string;
  score: number;
};

const KEY_OPEN_PRIMARY = "k";
const FETCH_DEBOUNCE_MS = 150;

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [hits, setHits] = useState<Hit[]>([]);
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === KEY_OPEN_PRIMARY) {
        e.preventDefault();
        setOpen((v) => !v);
      }
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  useEffect(() => {
    if (open) {
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [open]);

  useEffect(() => {
    if (!open || !query) {
      setHits([]);
      return;
    }
    const handle = setTimeout(async () => {
      try {
        const res = await fetch(`/api/palette?q=${encodeURIComponent(query)}`);
        if (res.ok) {
          const data = (await res.json()) as Hit[];
          setHits(data || []);
          setActive(0);
        }
      } catch {
        setHits([]);
      }
    }, FETCH_DEBOUNCE_MS);
    return () => clearTimeout(handle);
  }, [open, query]);

  const total = hits.length;
  const selected = useMemo(() => hits[active], [hits, active]);

  if (!open) {
    return <button data-testid="palette-trigger" onClick={() => setOpen(true)} style={{ display: "none" }}>open</button>;
  }

  return (
    <div
      role="dialog"
      aria-label="Command palette"
      data-testid="palette"
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(15,23,42,0.4)",
        zIndex: tokens.z.toast,
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "center",
        paddingTop: 80,
      }}
      onClick={() => setOpen(false)}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: tokens.color.surface,
          border: `1px solid ${tokens.color.border}`,
          borderRadius: tokens.radius.md,
          width: "min(640px, 90vw)",
          padding: tokens.space(3),
        }}
      >
        <input
          ref={inputRef}
          aria-label="search palette"
          data-testid="palette-input"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "ArrowDown") setActive((a) => Math.min(total - 1, a + 1));
            if (e.key === "ArrowUp") setActive((a) => Math.max(0, a - 1));
            if (e.key === "Enter" && selected?.router_path) {
              navigate(selected.router_path);
              setOpen(false);
            }
          }}
          placeholder="Search projects, capabilities, commands…"
          style={{
            width: "100%",
            padding: tokens.space(2),
            border: `1px solid ${tokens.color.border}`,
            borderRadius: tokens.radius.sm,
            fontSize: 14,
          }}
        />
        <ul
          role="listbox"
          data-testid="palette-list"
          style={{ listStyle: "none", padding: 0, margin: `${tokens.space(2)} 0 0`, maxHeight: 320, overflow: "auto" }}
        >
          {hits.map((h, i) => (
            <li
              key={`${h.source}-${h.title}-${i}`}
              role="option"
              aria-selected={i === active}
              data-testid={`palette-hit-${i}`}
              onMouseEnter={() => setActive(i)}
              onClick={() => {
                if (h.router_path) {
                  navigate(h.router_path);
                  setOpen(false);
                }
              }}
              style={{
                padding: tokens.space(2),
                borderRadius: tokens.radius.sm,
                background: i === active ? `${tokens.color.primary}10` : "transparent",
                cursor: "pointer",
                display: "flex",
                justifyContent: "space-between",
                gap: tokens.space(2),
              }}
            >
              <span>
                <strong>{h.title}</strong>
                <span style={{ color: tokens.color.textMuted, marginLeft: tokens.space(2), fontSize: 12 }}>{h.kind}</span>
              </span>
              <span style={{ color: tokens.color.textMuted, fontSize: 12 }}>{h.source}</span>
            </li>
          ))}
          {hits.length === 0 && (
            <li role="option" data-testid="palette-empty" style={{ color: tokens.color.textMuted, padding: tokens.space(2) }}>
              no matches
            </li>
          )}
        </ul>
      </div>
    </div>
  );
}
