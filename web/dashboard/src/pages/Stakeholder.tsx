import { Badge, Card } from "../ds";

type Item = {
  label: string;
  status: "ready" | "blocked" | "mocked" | "approval";
  note: string;
};

const ITEMS: Item[] = [
  { label: "Workspace + multi-project switching", status: "ready", note: "operator can register N projects and switch" },
  { label: "Capabilities install (agent/mcp/hook)", status: "ready", note: "deterministic discovery; install requires --yes" },
  { label: "Cleanup engine (worktrees, caches, leftovers)", status: "ready", note: "policy file required to delete anything" },
  { label: "Active run streaming", status: "mocked", note: "demo timeline; SSE wiring queued P30" },
  { label: "Design-to-product full state machine", status: "mocked", note: "feature toggle rules pending" },
  { label: "Memory promote/retire with evidence", status: "approval", note: "backend ok, UI in P30" },
  { label: "Stakeholder export PDF", status: "blocked", note: "needs report template + branding decision" },
];

function tone(status: Item["status"]) {
  switch (status) {
    case "ready":
      return "success" as const;
    case "blocked":
      return "danger" as const;
    case "mocked":
      return "warning" as const;
    case "approval":
      return "info" as const;
  }
}

export function StakeholderPage() {
  const groups = (["ready", "blocked", "mocked", "approval"] as const).map((status) => ({
    status,
    items: ITEMS.filter((it) => it.status === status),
  }));
  return (
    <div data-testid="page-stakeholder" style={{ display: "grid", gap: 16 }}>
      <div data-testid="stakeholder-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}>
        {groups.map((g) => (
          <Card key={g.status} title={`${g.status} (${g.items.length})`}>
            <ul data-testid={`stakeholder-${g.status}-list`} style={{ margin: 0, paddingLeft: 18 }}>
              {g.items.map((it, idx) => (
                <li key={idx} style={{ marginBottom: 6, fontSize: 13 }}>
                  <Badge tone={tone(g.status)} dot>{g.status}</Badge>
                  <strong style={{ marginLeft: 6 }}>{it.label}</strong>
                  <div style={{ color: "#64748B", marginLeft: 6 }}>{it.note}</div>
                </li>
              ))}
            </ul>
          </Card>
        ))}
      </div>
      <Card title="Next decision">
        <p data-testid="stakeholder-next-decision" style={{ margin: 0, fontSize: 14 }}>
          Wire <code>POST /api/actions/approve-changes</code> for memory promote/retire before exposing the dashboard to engineers outside the core team.
        </p>
      </Card>
    </div>
  );
}
