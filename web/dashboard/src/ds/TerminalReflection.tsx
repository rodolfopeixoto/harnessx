import { useEffect, useState } from "react";
import { TerminalReflectionService, type TerminalEntry } from "../lib/terminal";
import { tokens } from "./tokens";

type Props = {
  testId?: string;
  emptyMessage?: string;
  limit?: number;
};

const REFRESH_INTERVAL_MS = 1500;
const DEFAULT_LIMIT = 8;

export function TerminalReflection({ testId = "terminal-reflection", emptyMessage = "no commands yet", limit = DEFAULT_LIMIT }: Props) {
  const [entries, setEntries] = useState<TerminalEntry[]>([]);
  useEffect(() => {
    setEntries(TerminalReflectionService.list().slice(0, limit));
    const handle = setInterval(() => {
      setEntries(TerminalReflectionService.list().slice(0, limit));
    }, REFRESH_INTERVAL_MS);
    return () => clearInterval(handle);
  }, [limit]);

  return (
    <div
      data-testid={testId}
      style={{
        background: "#0F172A",
        color: "#E2E8F0",
        borderRadius: tokens.radius.md,
        padding: tokens.space(3),
        fontFamily: tokens.font.mono,
        fontSize: 12,
        lineHeight: 1.5,
        overflowX: "auto",
      }}
    >
      {entries.length === 0 && <div style={{ color: "#64748B" }}>{emptyMessage}</div>}
      {entries.map((e) => (
        <div key={e.id} data-testid="terminal-entry" style={{ display: "flex", gap: tokens.space(2) }}>
          <span style={{ color: e.status === "ok" ? "#34D399" : e.status === "failed" ? "#F87171" : "#FACC15" }}>
            {e.status === "ok" ? "✓" : e.status === "failed" ? "✗" : "…"}
          </span>
          <span>{e.command}</span>
        </div>
      ))}
    </div>
  );
}
