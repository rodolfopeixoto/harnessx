import { useState } from "react";
import { Badge, Card } from "../ds";
import { ActionService } from "../lib/actions";

const STEPS = [
  { id: "folder", label: "Choose folder", detail: "/Users/<you>/dev/<project>" },
  { id: "stack", label: "Detect stack", detail: "go · node · ruby · python · container" },
  { id: "agents", label: "Probe agents", detail: "claude · codex · gemini · kimi · fake" },
  { id: "sensors", label: "Probe sensors", detail: "go vet · go test · dashboard build · stack audit" },
  { id: "index", label: "Pick index mode", detail: "quick · standard · deep" },
  { id: "doctor", label: "Run doctor", detail: "harness doctor --plain" },
  { id: "done", label: "Done", detail: "harness project current" },
];

export function OnboardingPage() {
  const [step, setStep] = useState(0);
  const advance = () => {
    const next = Math.min(step + 1, STEPS.length - 1);
    setStep(next);
    ActionService.record("onboarding.advance", STEPS[next].id, `harness project import --yes`);
  };
  return (
    <div data-testid="page-onboarding" style={{ display: "grid", gap: 16 }}>
      <Card title="First run wizard">
        <ol data-testid="onboarding-steps" style={{ margin: 0, paddingLeft: 18 }}>
          {STEPS.map((s, i) => (
            <li key={s.id} style={{ marginBottom: 8, display: "flex", gap: 8, alignItems: "center" }}>
              <Badge tone={i < step ? "success" : i === step ? "primary" : "neutral"} dot>
                {i < step ? "done" : i === step ? "current" : "pending"}
              </Badge>
              <strong style={{ minWidth: 180 }}>{s.label}</strong>
              <span style={{ color: "#64748B", fontSize: 13 }}>{s.detail}</span>
            </li>
          ))}
        </ol>
        <div style={{ marginTop: 12 }}>
          <button data-testid="onboarding-next-action" onClick={advance} style={{ padding: "8px 14px", borderRadius: 6, background: "#4338CA", color: "white", border: "none", cursor: "pointer" }}>
            {step < STEPS.length - 1 ? "Next step" : "Restart"}
          </button>
        </div>
      </Card>
    </div>
  );
}
