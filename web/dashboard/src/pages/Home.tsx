import { useEffect, useState } from "react";
import { api } from "../api";
import { Card, MetricCard, TerminalReflection, Badge } from "../ds";

type WorkspaceProject = {
  Slug: string;
  DisplayName: string;
  RootPath: string;
  LastSeenAt?: string;
};

type HealthScore = {
  total: number;
  subsystems: Array<{ name: string; score: number; reason: string }>;
};

const SAMPLE_NEXT_ACTION = "Run `harness stack audit` to capture a fresh report.";

export function HomePage() {
  const [projects, setProjects] = useState<WorkspaceProject[]>([]);
  const [sessions, setSessions] = useState<Array<{ ID: string; Mode: string; Status: string; StartedAt?: string }>>([]);
  const [health, setHealth] = useState<HealthScore | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const res = await fetch("/api/workspace/projects");
        if (res.ok) setProjects(((await res.json()) as WorkspaceProject[]) ?? []);
      } catch {
        /* offline */
      }
      try {
        const data = await api.sessions(10);
        setSessions((data as typeof sessions) ?? []);
      } catch {
        /* offline */
      }
      try {
        const res = await fetch("/api/health/score");
        if (res.ok) setHealth((await res.json()) as HealthScore);
      } catch {
        /* offline */
      }
    })();
  }, []);

  const activeProject = projects[0];
  return (
    <div data-testid="page-home" style={{ display: "grid", gap: 16 }}>
      <div data-testid="workspace-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}>
        <MetricCard
          testId="workspace-projects-card"
          label="Projects"
          value={projects.length}
          hint={activeProject ? `active: ${activeProject.DisplayName}` : "run `harness project add <path>` to register"}
          tone="primary"
        />
        <MetricCard
          testId="recent-runs-count-card"
          label="Recent runs"
          value={sessions.length}
          hint="last 10 sessions"
          tone={sessions.length > 0 ? "success" : "neutral"}
        />
        <MetricCard
          testId="health-score-card"
          label="Health score"
          value={health ? `${health.total}/100` : "—"}
          hint={health ? `${health.subsystems.length} subsystems weighted` : "/api/health/score not yet computed"}
          tone={health && health.total > 70 ? "success" : "warning"}
        />
        <MetricCard
          testId="next-action-card"
          label="Next action"
          value="Audit"
          hint={SAMPLE_NEXT_ACTION}
          tone="info"
        />
      </div>

      <Card title="Recent runs">
        <div data-testid="recent-runs">
          {sessions.length === 0 ? (
            <p style={{ margin: 0, color: "#64748B" }}>no sessions recorded — start one with <code>harness feature ...</code></p>
          ) : (
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={{ textAlign: "left", fontSize: 11, color: "#64748B", padding: "8px 0" }}>id</th>
                  <th style={{ textAlign: "left", fontSize: 11, color: "#64748B", padding: "8px 0" }}>mode</th>
                  <th style={{ textAlign: "left", fontSize: 11, color: "#64748B", padding: "8px 0" }}>status</th>
                  <th style={{ textAlign: "left", fontSize: 11, color: "#64748B", padding: "8px 0" }}>started</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map((row) => (
                  <tr key={row.ID} data-testid="recent-runs-row">
                    <td style={{ padding: "6px 0", fontSize: 13 }}>{row.ID.slice(0, 12)}…</td>
                    <td style={{ padding: "6px 0", fontSize: 13 }}>
                      <Badge tone="info" dot>{row.Mode}</Badge>
                    </td>
                    <td style={{ padding: "6px 0", fontSize: 13 }}>
                      <Badge tone={row.Status === "succeeded" ? "success" : row.Status === "failed" ? "danger" : "neutral"} dot>{row.Status}</Badge>
                    </td>
                    <td style={{ padding: "6px 0", fontSize: 12, color: "#64748B" }}>{row.StartedAt ?? ""}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </Card>

      <Card title="Terminal reflection">
        <TerminalReflection testId="terminal-reflection" />
      </Card>
    </div>
  );
}
