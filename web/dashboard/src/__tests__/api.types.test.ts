import { describe, it, expect } from "vitest";
import type { Session, SensorRow, Cost, RuntimeRow } from "../api";

describe("api type contracts", () => {
  it("Session accepts canonical fields", () => {
    const s: Session = {
      ID: "01ABC",
      ProjectPath: "/tmp/x",
      Mode: "feature",
      Status: "running",
    };
    expect(s.ID).toBe("01ABC");
    expect(s.Mode).toBe("feature");
  });

  it("SensorRow narrows status to expected values", () => {
    const r: SensorRow = {
      id: 1,
      run_id: "01ABC",
      sensor: "secrets_scan",
      status: "passed",
      duration_ms: 42,
      created_at: "2026-06-16",
    };
    expect(["passed", "failed", "skipped"]).toContain(r.status);
  });

  it("Cost.by_agent items have cost_usd + token fields", () => {
    const c: Cost = {
      total_usd: 0.05,
      by_agent: [
        { agent: "claude", cost_usd: 0.05, input_tokens: 100, output_tokens: 200 },
      ],
    };
    const first = c.by_agent?.[0];
    expect(first?.cost_usd).toBe(0.05);
    expect(first?.input_tokens).toBe(100);
  });

  it("RuntimeRow boolean fields work as toggles", () => {
    const r: RuntimeRow = { id: "docker", binary: "docker", available: true, selected: false };
    expect(r.available).toBe(true);
    expect(r.selected).toBe(false);
  });
});
