import { useMemo, useState } from "react";
import { Badge, Card, MetricCard, TerminalReflection } from "../ds";
import { ActionService } from "../lib/actions";

const MODES = [
  { id: "feature", label: "Feature", hint: "Spec → build → test a new capability." },
  { id: "bugfix", label: "Bugfix", hint: "Reproduce, fix, prove with tests." },
  { id: "design-to-product", label: "Design to Product", hint: "ZIP/image → React parity → roadmap." },
  { id: "ask", label: "Ask", hint: "Read-only Q&A, no file changes." },
  { id: "audit", label: "Audit", hint: "Deterministic visual + functional sweep." },
];

const SAMPLE_PROMPT = "Add a /healthz endpoint that returns 200 and the build sha.";

function detectMode(prompt: string): { id: string; confidence: number } {
  const lower = prompt.toLowerCase();
  if (lower.includes("zip") || lower.includes("design") || lower.includes("figma")) return { id: "design-to-product", confidence: 94 };
  if (lower.includes("fix") || lower.includes("bug")) return { id: "bugfix", confidence: 88 };
  if (lower.includes("explain") || lower.includes("what is")) return { id: "ask", confidence: 78 };
  if (lower.includes("audit") || lower.includes("review")) return { id: "audit", confidence: 80 };
  return { id: "feature", confidence: 74 };
}

export function CommandPage() {
  const [prompt, setPrompt] = useState(SAMPLE_PROMPT);
  const detection = useMemo(() => detectMode(prompt), [prompt]);
  const cmd = useMemo(() => {
    const escaped = prompt.replace(/"/g, '\\"');
    return `harness ${detection.id} "${escaped}" --yes`;
  }, [detection.id, prompt]);

  const start = () => {
    ActionService.record("command.start", `${detection.id}: ${prompt.slice(0, 40)}`, cmd);
  };
  const plan = () => {
    ActionService.record("command.plan", `${detection.id}: ${prompt.slice(0, 40)}`, cmd.replace("--yes", "--plan-only"));
  };

  return (
    <div data-testid="page-command" style={{ display: "grid", gap: 16 }}>
      <Card title="Natural prompt">
        <textarea
          data-testid="command-prompt"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={4}
          style={{ width: "100%", padding: 12, fontSize: 14, border: "1px solid #E4E4E7", borderRadius: 8, fontFamily: "inherit" }}
          aria-label="prompt"
        />
        <div data-testid="command-detection" style={{ marginTop: 12, display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
          <MetricCard testId="command-mode-card" label="Detected mode" value={detection.id} hint={`${detection.confidence}% confidence`} tone="primary" />
          <MetricCard testId="command-cost-card" label="Cost estimate" value="≈ $0.04" hint="claude opus 4.7 · 12k tokens" tone="info" />
          <MetricCard testId="command-risk-card" label="Risk" value="medium" hint="touches backend + frontend" tone="warning" />
        </div>
        <div data-testid="command-actions" style={{ marginTop: 12, display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button data-testid="command-plan-action" onClick={plan} style={{ padding: "8px 14px", borderRadius: 6, border: "1px solid #4338CA", color: "#4338CA", background: "white", cursor: "pointer" }}>Plan only</button>
          <button data-testid="command-start-action" onClick={start} style={{ padding: "8px 14px", borderRadius: 6, border: "1px solid #4338CA", color: "white", background: "#4338CA", cursor: "pointer" }}>Start run</button>
        </div>
      </Card>

      <Card title="Available modes">
        <div data-testid="command-modes" style={{ display: "grid", gap: 8 }}>
          {MODES.map((m) => (
            <div key={m.id} style={{ display: "flex", gap: 12, alignItems: "center" }}>
              <Badge tone={m.id === detection.id ? "primary" : "neutral"} dot>{m.id}</Badge>
              <strong style={{ minWidth: 140 }}>{m.label}</strong>
              <span style={{ color: "#64748B", fontSize: 13 }}>{m.hint}</span>
            </div>
          ))}
        </div>
      </Card>

      <Card title="Terminal reflection">
        <TerminalReflection />
      </Card>
    </div>
  );
}
