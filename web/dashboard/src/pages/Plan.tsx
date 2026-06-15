import { useState } from "react";
import { Badge, Card, MetricCard } from "../ds";
import { ActionService } from "../lib/actions";

const SAMPLE_FILES = [
  { path: "internal/adapters/http/handlers.go", action: "modify", reason: "add /healthz handler" },
  { path: "internal/adapters/http/server.go", action: "modify", reason: "register route" },
  { path: "internal/adapters/http/handlers_test.go", action: "modify", reason: "happy path + 500 path" },
];

const SAMPLE_RISKS = [
  "no auth on /healthz — confirm public exposure intended",
  "no rate limit — fine for k8s readiness probes",
];

const SAMPLE_SENSORS = ["go_vet", "go_test", "dashboard-build", "stack-audit"];

export function PlanPage() {
  const [decision, setDecision] = useState<string>("");

  const approve = () => {
    setDecision("approved");
    ActionService.record("plan.approve", "demo-plan", "harness plan approve demo-plan");
  };
  const request = () => {
    setDecision("changes_requested");
    ActionService.record("plan.request-changes", "demo-plan", "harness plan request-changes demo-plan");
  };
  const cheaper = () => {
    ActionService.record("plan.regen-cheaper", "demo-plan", "harness plan regenerate --mode cheap demo-plan");
  };
  const safer = () => {
    ActionService.record("plan.regen-safer", "demo-plan", "harness plan regenerate --mode safe demo-plan");
  };

  return (
    <div data-testid="page-plan" style={{ display: "grid", gap: 16 }}>
      <div data-testid="plan-demo-banner" style={{ background: "#EFF6FF", border: "1px solid #BFDBFE", padding: 10, borderRadius: 8, fontSize: 12 }}>
        Demo plan. Wire `POST /api/actions/approve-plan` to persist real decisions.
      </div>

      <Card title="Plan summary">
        <div data-testid="plan-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
          <MetricCard testId="plan-files-card" label="Files touched" value={SAMPLE_FILES.length} tone="primary" />
          <MetricCard testId="plan-sensors-card" label="Sensors to run" value={SAMPLE_SENSORS.length} tone="info" />
          <MetricCard testId="plan-risks-card" label="Risks flagged" value={SAMPLE_RISKS.length} tone={SAMPLE_RISKS.length > 0 ? "warning" : "success"} />
          <MetricCard testId="plan-decision-card" label="Decision" value={decision || "pending"} tone={decision === "approved" ? "success" : decision === "changes_requested" ? "danger" : "neutral"} />
        </div>
      </Card>

      <Card title="Files">
        <div data-testid="plan-files" style={{ display: "grid", gap: 6 }}>
          {SAMPLE_FILES.map((f) => (
            <div key={f.path} style={{ display: "grid", gridTemplateColumns: "100px 1fr 2fr", gap: 8, fontSize: 13 }}>
              <Badge tone={f.action === "create" ? "success" : f.action === "delete" ? "danger" : "info"} dot>{f.action}</Badge>
              <code>{f.path}</code>
              <span style={{ color: "#64748B" }}>{f.reason}</span>
            </div>
          ))}
        </div>
      </Card>

      <Card title="Risks">
        <ul data-testid="plan-risks" style={{ margin: 0, paddingLeft: 18 }}>
          {SAMPLE_RISKS.map((r, i) => (
            <li key={i} style={{ fontSize: 13, marginBottom: 4 }}>{r}</li>
          ))}
        </ul>
      </Card>

      <Card title="Sensors that will run">
        <div data-testid="plan-sensors" style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
          {SAMPLE_SENSORS.map((s) => (
            <Badge key={s} tone="info">{s}</Badge>
          ))}
        </div>
      </Card>

      <Card title="Decisions">
        <div data-testid="plan-actions" style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
          <button data-testid="plan-approve-action" onClick={approve} style={{ padding: "8px 14px", borderRadius: 6, background: "#15803D", color: "white", border: "none", cursor: "pointer" }}>Approve plan</button>
          <button data-testid="plan-request-action" onClick={request} style={{ padding: "8px 14px", borderRadius: 6, background: "#B91C1C", color: "white", border: "none", cursor: "pointer" }}>Request changes</button>
          <button data-testid="plan-cheaper-action" onClick={cheaper} style={{ padding: "8px 14px", borderRadius: 6, background: "white", color: "#4338CA", border: "1px solid #4338CA", cursor: "pointer" }}>Cheaper plan</button>
          <button data-testid="plan-safer-action" onClick={safer} style={{ padding: "8px 14px", borderRadius: 6, background: "white", color: "#4338CA", border: "1px solid #4338CA", cursor: "pointer" }}>Safer plan</button>
        </div>
      </Card>
    </div>
  );
}
