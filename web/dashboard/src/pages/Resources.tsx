import { Badge, Card, MetricCard } from "../ds";

type Resource = {
  category: string;
  label: string;
  usage: string;
  tone: "success" | "warning" | "danger" | "neutral";
};

const SAMPLE: Resource[] = [
  { category: "Containers", label: "0 running · 0 stopped", usage: "no docker on PATH", tone: "neutral" },
  { category: "Caches", label: "node_modules · 268 MB", usage: "untouched 12 days", tone: "warning" },
  { category: "Worktrees", label: "0 stale", usage: "git worktree list clean", tone: "success" },
  { category: "Artifacts", label: ".harness/artifacts · 12 MB", usage: "specs + reports + memory", tone: "neutral" },
  { category: "Logs", label: ".harness/logs · 4 MB", usage: "5 rotated files", tone: "neutral" },
];

export function ResourcesPage() {
  return (
    <div data-testid="page-resources" style={{ display: "grid", gap: 16 }}>
      <div data-testid="resources-demo-banner" style={{ background: "#EFF6FF", border: "1px solid #BFDBFE", padding: 10, borderRadius: 8, fontSize: 12 }}>
        Demo snapshot. Wire <code>GET /api/resources</code> to surface live usage (P30). Use <code>harness cleanup scan</code> for the real worktree/cache inventory.
      </div>
      <div data-testid="resources-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))" }}>
        <MetricCard testId="resources-cpu-card" label="CPU" value="2 cores" hint="docker.cpus=2.0" tone="info" />
        <MetricCard testId="resources-memory-card" label="Memory" value="2 GB" hint="docker mem_limit" tone="info" />
        <MetricCard testId="resources-disk-card" label="Disk" value="284 MB" hint="across categories below" tone="warning" />
        <MetricCard testId="resources-deps-card" label="Dependency size" value="178 MB" hint="lockfile total" tone="neutral" />
      </div>
      <Card title="Categories">
        <div data-testid="resources-categories" style={{ display: "grid", gap: 8 }}>
          {SAMPLE.map((r) => (
            <div key={r.category} style={{ display: "grid", gridTemplateColumns: "160px 1fr 1fr", gap: 12, alignItems: "center", fontSize: 13 }}>
              <Badge tone={r.tone === "warning" ? "warning" : r.tone === "danger" ? "danger" : r.tone === "success" ? "success" : "neutral"} dot>{r.category}</Badge>
              <strong>{r.label}</strong>
              <span style={{ color: "#64748B" }}>{r.usage}</span>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
