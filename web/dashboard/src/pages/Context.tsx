import { Badge, Card, DataExplorer, type Column, MetricCard, PathCell } from "../ds";

type ContextFile = {
  path: string;
  source: string;
  reason: string;
  tokens: number;
  included: boolean;
};

const SAMPLE: ContextFile[] = [
  { path: "internal/adapters/http/handlers.go", source: "git", reason: "git_status modified", tokens: 612, included: true },
  { path: "internal/adapters/http/server.go", source: "lsp", reason: "imports handlers.go", tokens: 318, included: true },
  { path: "internal/adapters/http/handlers_test.go", source: "test_map", reason: "covers handlers", tokens: 540, included: true },
  { path: "internal/platform/constants/constants.go", source: "ripgrep", reason: "matches `health`", tokens: 244, included: true },
  { path: "web/dashboard/dist/assets/index.js", source: "git", reason: "build artefact", tokens: 18120, included: false },
  { path: "docs/design-handoff-v2/harness-cli-ui/project/styles.css", source: "git", reason: "vendored design handoff", tokens: 9700, included: false },
];

export function ContextPage() {
  const included = SAMPLE.filter((f) => f.included);
  const rejected = SAMPLE.filter((f) => !f.included);
  const totalTokens = included.reduce((sum, f) => sum + f.tokens, 0);
  const fullRepoEstimate = SAMPLE.reduce((sum, f) => sum + f.tokens, 0);
  const saved = fullRepoEstimate - totalTokens;

  const columns: Column<ContextFile>[] = [
    { id: "status", header: "STATUS", render: (f) => <Badge tone={f.included ? "success" : "neutral"} dot>{f.included ? "included" : "rejected"}</Badge> },
    { id: "source", header: "SOURCE", render: (f) => <Badge tone="info">{f.source}</Badge> },
    { id: "path", header: "PATH", render: (f) => <PathCell path={f.path} testId="context-path-cell" /> },
    { id: "tokens", header: "TOKENS", render: (f) => f.tokens.toLocaleString() },
    { id: "reason", header: "REASON", render: (f) => f.reason },
  ];

  return (
    <div data-testid="page-context" style={{ display: "grid", gap: 16 }}>
      <div data-testid="context-demo-banner" style={{ background: "#EFF6FF", border: "1px solid #BFDBFE", padding: 10, borderRadius: 8, fontSize: 12 }}>
        Demo context pack. Wire <code>GET /api/context</code> to surface the live pack (P30).
      </div>
      <div data-testid="context-summary" style={{ display: "grid", gap: 12, gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))" }}>
        <MetricCard testId="context-included-card" label="Included files" value={included.length} hint="passed through providers" tone="primary" />
        <MetricCard testId="context-rejected-card" label="Rejected files" value={rejected.length} hint="excluded by reason" tone="warning" />
        <MetricCard testId="context-tokens-card" label="Pack tokens" value={totalTokens.toLocaleString()} hint={`vs ${fullRepoEstimate.toLocaleString()} full-repo`} tone="info" />
        <MetricCard testId="context-saved-card" label="Tokens saved" value={saved.toLocaleString()} hint={`${Math.round((saved / fullRepoEstimate) * 100)}% reduction`} tone="success" />
      </div>
      <Card title="Files">
        <div data-testid="context-explorer">
          <DataExplorer items={SAMPLE as unknown as Record<string, unknown>[]} columns={columns as unknown as Column<Record<string, unknown>>[]} searchKeys={["path", "reason"]} pageSize={20} />
        </div>
      </Card>
    </div>
  );
}
