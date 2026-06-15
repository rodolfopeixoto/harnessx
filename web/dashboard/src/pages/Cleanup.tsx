import { useEffect, useState } from "react";
import { Card, DataExplorer, type Column, PathCell, Badge } from "../ds";

type Finding = {
  Kind: string;
  Path: string;
  Risk: "low" | "medium" | "high" | string;
  Reason?: string;
  SizeBytes?: number;
} & Record<string, unknown>;

const KB = 1024;
const MB = 1024 * 1024;
const GB = 1024 * 1024 * 1024;

function formatBytes(n?: number): string {
  if (!n) return "—";
  if (n > GB) return `${(n / GB).toFixed(2)} GB`;
  if (n > MB) return `${(n / MB).toFixed(2)} MB`;
  if (n > KB) return `${(n / KB).toFixed(2)} KB`;
  return `${n} B`;
}

function riskTone(risk: string) {
  switch (risk) {
    case "high":
      return "danger" as const;
    case "medium":
      return "warning" as const;
    default:
      return "neutral" as const;
  }
}

export function CleanupPage() {
  const [findings, setFindings] = useState<Finding[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const res = await fetch("/api/cleanup/scan");
        if (!res.ok) {
          setError(`API ${res.status}`);
          return;
        }
        setFindings(((await res.json()) as Finding[]) ?? []);
      } catch (err) {
        setError(String(err));
      }
    })();
  }, []);

  const totalSize = findings.reduce((sum, f) => sum + (f.SizeBytes ?? 0), 0);

  const columns: Column<Finding>[] = [
    {
      id: "risk",
      header: "RISK",
      render: (f) => (
        <Badge tone={riskTone(f.Risk as string)} dot>
          {f.Risk}
        </Badge>
      ),
    },
    { id: "kind", header: "KIND", render: (f) => f.Kind },
    { id: "size", header: "SIZE", render: (f) => formatBytes(f.SizeBytes) },
    { id: "path", header: "PATH", render: (f) => <PathCell path={f.Path} testId="cleanup-path-cell" /> },
    { id: "reason", header: "REASON", render: (f) => f.Reason ?? "—" },
  ];

  return (
    <div data-testid="page-cleanup" style={{ display: "grid", gap: 16 }}>
      <Card title="Cleanup scan">
        <div data-testid="cleanup-plan-banner" style={{ background: "#FEF9C3", border: "1px solid #FACC15", padding: 12, borderRadius: 8, marginBottom: 12, fontSize: 13 }}>
          Scan-only. <code>harness cleanup apply --policy &lt;file&gt;</code> required to delete anything.
        </div>
        <div data-testid="cleanup-summary" style={{ marginBottom: 12, fontSize: 13 }}>
          {findings.length} candidates · total reclaimable {formatBytes(totalSize)}
        </div>
        <div data-testid="cleanup-explorer">
          <DataExplorer
            items={findings}
            columns={columns}
            searchKeys={["Kind", "Path", "Reason"]}
            pageSize={25}
          />
        </div>
        {error && (
          <p role="alert" style={{ color: "#B91C1C", marginTop: 12 }}>
            {error}
          </p>
        )}
      </Card>
    </div>
  );
}
