import { useEffect, useMemo, useState } from "react";
import { Badge, Card, DataExplorer, type Column, PathCell, Tabs, type Tab } from "../ds";
import { ActionService } from "../lib/actions";

type Capability = {
  kind: string;
  name: string;
  version?: string;
  source?: string;
  status: string;
  description?: string;
  config_path?: string;
} & Record<string, unknown>;

type McpServer = {
  name: string;
  source: string;
  transport: string;
  risk: string;
  config_path: string;
} & Record<string, unknown>;

type Hook = {
  name: string;
  source: string;
  event: string;
  scope: string;
  status: string;
  risk: string;
  config_path: string;
} & Record<string, unknown>;

const KIND_LABELS: Record<string, string> = {
  agent: "Agents",
  mcp: "MCPs",
  hook: "Hooks",
  sensor: "Sensors",
  skill: "Skills",
  context: "Context Providers",
  resource: "Resource Providers",
  plugin: "Plugins",
};

function statusTone(status: string) {
  switch (status) {
    case "installed":
    case "enabled":
    case "configured":
      return "success" as const;
    case "failed":
    case "stale":
      return "danger" as const;
    case "detected":
    case "not_installed":
      return "neutral" as const;
    default:
      return "info" as const;
  }
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

export function CatalogPage() {
  const [kinds, setKinds] = useState<string[]>([]);
  const [items, setItems] = useState<Capability[]>([]);
  const [mcps, setMcps] = useState<McpServer[]>([]);
  const [hooks, setHooks] = useState<Hook[]>([]);

  useEffect(() => {
    void (async () => {
      await Promise.all([
        fetch("/api/catalog/kinds").then(async (r) => { if (r.ok) setKinds(((await r.json()) as string[]) ?? []); }).catch(() => undefined),
        fetch("/api/catalog/items").then(async (r) => { if (r.ok) setItems(((await r.json()) as Capability[]) ?? []); }).catch(() => undefined),
        fetch("/api/mcps").then(async (r) => { if (r.ok) setMcps(((await r.json()) as McpServer[]) ?? []); }).catch(() => undefined),
        fetch("/api/hooks").then(async (r) => { if (r.ok) setHooks(((await r.json()) as Hook[]) ?? []); }).catch(() => undefined),
      ]);
    })();
  }, []);

  const catalogColumns: Column<Capability>[] = [
    { id: "name", header: "NAME", render: (r) => <strong>{r.name}</strong> },
    { id: "status", header: "STATUS", render: (r) => <Badge tone={statusTone(r.status)} dot>{r.status}</Badge> },
    { id: "source", header: "SOURCE", render: (r) => r.source ?? "—" },
    { id: "version", header: "VERSION", render: (r) => r.version ?? "—" },
    { id: "desc", header: "DESCRIPTION", render: (r) => r.description ?? "—" },
  ];

  const mcpColumns: Column<McpServer>[] = [
    { id: "name", header: "NAME", render: (r) => <strong>{r.name}</strong> },
    { id: "source", header: "SOURCE", render: (r) => r.source },
    { id: "transport", header: "TRANSPORT", render: (r) => <Badge tone="info">{r.transport}</Badge> },
    { id: "risk", header: "RISK", render: (r) => <Badge tone={riskTone(r.risk)} dot>{r.risk}</Badge> },
    { id: "path", header: "PATH", render: (r) => <PathCell path={r.config_path} testId="mcp-path-cell" /> },
  ];

  const hookColumns: Column<Hook>[] = [
    { id: "name", header: "NAME", render: (r) => <strong>{r.name}</strong> },
    { id: "event", header: "EVENT", render: (r) => <Badge tone="info">{r.event}</Badge> },
    { id: "source", header: "SOURCE", render: (r) => r.source },
    { id: "scope", header: "SCOPE", render: (r) => r.scope },
    { id: "status", header: "STATUS", render: (r) => <Badge tone={r.status === "enabled" ? "success" : "neutral"} dot>{r.status}</Badge> },
    { id: "risk", header: "RISK", render: (r) => <Badge tone={riskTone(r.risk)} dot>{r.risk}</Badge> },
    { id: "path", header: "PATH", render: (r) => <PathCell path={r.config_path} testId="hook-path-cell" /> },
  ];

  const tabs: Tab[] = useMemo(() => {
    const presentKinds = kinds.length > 0 ? kinds : Object.keys(KIND_LABELS);
    return presentKinds.map((kind) => {
      if (kind === "mcp") {
        return {
          id: kind,
          label: KIND_LABELS[kind],
          render: () => (
            <div data-testid="mcp-tab">
              <DataExplorer
                items={mcps}
                columns={mcpColumns}
                searchKeys={["name", "source", "transport"]}
                pageSize={20}
                onInspect={(row) => {
                  ActionService.record("mcp.inspect", `${row.source}/${row.name}`, `harness mcp scan && cat ${row.config_path}`);
                }}
              />
            </div>
          ),
        };
      }
      if (kind === "hook") {
        return {
          id: kind,
          label: KIND_LABELS[kind],
          render: () => (
            <div data-testid="hook-tab">
              <DataExplorer
                items={hooks}
                columns={hookColumns}
                searchKeys={["name", "event", "source"]}
                pageSize={20}
                onInspect={(row) => {
                  ActionService.record("hook.inspect", `${row.source}/${row.name}`, `harness hook scan && cat ${row.config_path}`);
                }}
              />
            </div>
          ),
        };
      }
      const filtered = items.filter((i) => i.kind === kind);
      return {
        id: kind,
        label: KIND_LABELS[kind] ?? kind,
        render: () => (
          <div data-testid={`${kind}-tab`}>
            <DataExplorer
              items={filtered}
              columns={catalogColumns}
              searchKeys={["name", "description", "status"]}
              pageSize={20}
              onInspect={(row) => {
                ActionService.record("catalog.inspect", `${row.kind}/${row.name}`, `harness catalog show ${row.kind} ${row.name}`);
              }}
            />
          </div>
        ),
      };
    });
  }, [items, kinds, mcps, hooks]);

  return (
    <div data-testid="page-catalog" style={{ display: "grid", gap: 16 }}>
      <Card title="Capabilities">
        <div data-testid="capabilities-tabs">
          <Tabs tabs={tabs} />
        </div>
      </Card>
    </div>
  );
}
