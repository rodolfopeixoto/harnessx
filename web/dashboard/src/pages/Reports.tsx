import { Badge, Card, DataExplorer, type Column, MetricCard } from "../ds";

type Report = {
  id: string;
  kind: string;
  project: string;
  agent: string;
  cost: string;
  status: string;
  date: string;
};

const SAMPLE: Report[] = [
  { id: "rep-014", kind: "stack-audit", project: "harnessx", agent: "operator", cost: "—", status: "passed", date: "2026-06-14 23:08" },
  { id: "rep-013", kind: "design-to-product", project: "harnessx", agent: "claude", cost: "$0.18", status: "succeeded", date: "2026-06-14 22:30" },
  { id: "rep-012", kind: "security-audit", project: "harnessx", agent: "operator", cost: "—", status: "succeeded", date: "2026-06-14 18:11" },
  { id: "rep-011", kind: "feature", project: "harnessx", agent: "codex", cost: "$0.32", status: "succeeded", date: "2026-06-13 16:42" },
  { id: "rep-010", kind: "bugfix", project: "harnessx", agent: "claude", cost: "$0.07", status: "failed", date: "2026-06-13 14:10" },
];

function tone(status: string) {
  switch (status) {
    case "passed":
    case "succeeded":
      return "success" as const;
    case "failed":
      return "danger" as const;
    default:
      return "info" as const;
  }
}

export function ReportsPage() {
  const succeeded = SAMPLE.filter((r) => r.status === "succeeded" || r.status === "passed").length;
  const columns: Column<Report>[] = [
    { id: "id", header: "ID", render: (r) => <code>{r.id}</code> },
    { id: "kind", header: "KIND", render: (r) => <Badge tone="info">{r.kind}</Badge> },
    { id: "project", header: "PROJECT", render: (r) => r.project },
    { id: "agent", header: "AGENT", render: (r) => r.agent },
    { id: "cost", header: "COST", render: (r) => r.cost },
    { id: "status", header: "STATUS", render: (r) => <Badge tone={tone(r.status)} dot>{r.status}</Badge> },
    { id: "date", header: "DATE", render: (r) => r.date },
  ];
  return (
    <div data-testid="page-reports" style={{ display: "grid", gap: 16 }}>
      <div data-testid="reports-demo-banner" style={{ background: "#EFF6FF", border: "1px solid #BFDBFE", padding: 10, borderRadius: 8, fontSize: 12 }}>
        Demo archive. Wire <code>GET /api/reports</code> to list real run reports (P30).
      </div>
      <div data-testid="reports-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
        <MetricCard testId="reports-total-card" label="Reports" value={SAMPLE.length} tone="primary" />
        <MetricCard testId="reports-succeeded-card" label="Succeeded" value={succeeded} tone="success" />
        <MetricCard testId="reports-cost-card" label="Cost (sample)" value="$0.57" hint="last 5 runs" tone="info" />
      </div>
      <Card title="Reports">
        <div data-testid="reports-explorer">
          <DataExplorer items={SAMPLE as unknown as Record<string, unknown>[]} columns={columns as unknown as Column<Record<string, unknown>>[]} searchKeys={["id", "kind", "project", "agent", "status"]} pageSize={25} />
        </div>
      </Card>
    </div>
  );
}
