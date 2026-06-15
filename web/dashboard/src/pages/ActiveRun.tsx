import { Badge, Card, MetricCard, TerminalReflection } from "../ds";
import { ActionService } from "../lib/actions";

type Stage = {
  id: string;
  name: string;
  status: "done" | "active" | "pending" | "failed";
  detail: string;
};

const SAMPLE_STAGES: Stage[] = [
  { id: "spec", name: "Spec drafted", status: "done", detail: "feature-12-healthz.spec.md" },
  { id: "plan", name: "Plan approved", status: "done", detail: "3 files · 4 sensors" },
  { id: "context", name: "Context built", status: "done", detail: "12 files · 3 421 tokens" },
  { id: "execute", name: "Executing", status: "active", detail: "claude opus 4.7" },
  { id: "sensors", name: "Sensors", status: "pending", detail: "queued" },
  { id: "report", name: "Report", status: "pending", detail: "" },
];

const SAMPLE_EVENTS = [
  "12:14:01  spec.created  feature-12-healthz",
  "12:14:08  plan.approved by operator",
  "12:14:12  context.built  files=12 tokens=3421",
  "12:14:30  agent.start  claude opus 4.7",
  "12:14:42  diff.proposed handlers.go +18 -0",
];

export function ActiveRunPage() {
  const approve = () => ActionService.record("run.approve-changes", "demo-run", "harness run approve demo-run --artifact diff_123");
  const reqChanges = () => ActionService.record("run.request-changes", "demo-run", "harness run request-changes demo-run");
  const fallback = () => ActionService.record("run.fallback", "demo-run", "harness run fallback demo-run --next codex");
  const stop = () => ActionService.record("run.stop", "demo-run", "harness run stop demo-run");

  return (
    <div data-testid="page-activerun" style={{ display: "grid", gap: 16 }}>
      <div data-testid="activerun-demo-banner" style={{ background: "#EFF6FF", border: "1px solid #BFDBFE", padding: 10, borderRadius: 8, fontSize: 12 }}>
        Demo run. Live runs stream over SSE from <code>/api/events?run_id=...</code> (P30).
      </div>

      <div data-testid="activerun-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
        <MetricCard testId="run-stage-card" label="Stage" value="execute" hint="claude opus 4.7" tone="primary" />
        <MetricCard testId="run-tokens-card" label="Tokens" value="3 421" hint="in 2 412 / out 1 009" tone="info" />
        <MetricCard testId="run-cost-card" label="Cost" value="$0.041" hint="≈ $0.04 estimated" tone="info" />
        <MetricCard testId="run-latency-card" label="Latency" value="14 s" hint="agent IO so far" tone="neutral" />
      </div>

      <Card title="Timeline">
        <ol data-testid="run-timeline" style={{ margin: 0, paddingLeft: 18 }}>
          {SAMPLE_STAGES.map((s) => (
            <li key={s.id} style={{ marginBottom: 8, display: "flex", gap: 8, alignItems: "center" }}>
              <Badge tone={s.status === "done" ? "success" : s.status === "active" ? "primary" : s.status === "failed" ? "danger" : "neutral"} dot>
                {s.status}
              </Badge>
              <strong style={{ minWidth: 160 }}>{s.name}</strong>
              <span style={{ color: "#64748B", fontSize: 13 }}>{s.detail}</span>
            </li>
          ))}
        </ol>
      </Card>

      <Card title="Event stream">
        <ul data-testid="run-event-stream" style={{ margin: 0, paddingLeft: 18, fontFamily: "ui-monospace, monospace", fontSize: 12, lineHeight: 1.7 }}>
          {SAMPLE_EVENTS.map((e, i) => <li key={i}>{e}</li>)}
        </ul>
      </Card>

      <Card title="Actions">
        <div data-testid="run-actions" style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button data-testid="approve-changes-action" onClick={approve} style={{ padding: "8px 14px", borderRadius: 6, background: "#15803D", color: "white", border: "none", cursor: "pointer" }}>Approve changes</button>
          <button data-testid="request-changes-action" onClick={reqChanges} style={{ padding: "8px 14px", borderRadius: 6, background: "white", color: "#B45309", border: "1px solid #B45309", cursor: "pointer" }}>Request changes</button>
          <button data-testid="fallback-action" onClick={fallback} style={{ padding: "8px 14px", borderRadius: 6, background: "white", color: "#4338CA", border: "1px solid #4338CA", cursor: "pointer" }}>Run fallback</button>
          <button data-testid="stop-action" onClick={stop} style={{ padding: "8px 14px", borderRadius: 6, background: "white", color: "#B91C1C", border: "1px solid #B91C1C", cursor: "pointer" }}>Stop</button>
        </div>
      </Card>

      <Card title="Terminal reflection">
        <TerminalReflection testId="terminal-reflection" />
      </Card>
    </div>
  );
}
