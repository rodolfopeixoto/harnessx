import { useEffect, useMemo, useState } from "react";
import { Card, DataExplorer, type Column, Badge, Tabs, type Tab } from "../ds";
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

export function CatalogPage() {
  const [kinds, setKinds] = useState<string[]>([]);
  const [items, setItems] = useState<Capability[]>([]);

  useEffect(() => {
    void (async () => {
      try {
        const k = await fetch("/api/catalog/kinds");
        if (k.ok) setKinds(((await k.json()) as string[]) ?? []);
      } catch {
        /* offline */
      }
      try {
        const i = await fetch("/api/catalog/items");
        if (i.ok) setItems(((await i.json()) as Capability[]) ?? []);
      } catch {
        /* offline */
      }
    })();
  }, []);

  const columns: Column<Capability>[] = [
    { id: "name", header: "NAME", render: (r) => <strong>{r.name}</strong> },
    {
      id: "status",
      header: "STATUS",
      render: (r) => (
        <Badge tone={statusTone(r.status)} dot>
          {r.status}
        </Badge>
      ),
    },
    { id: "source", header: "SOURCE", render: (r) => r.source ?? "—" },
    { id: "version", header: "VERSION", render: (r) => r.version ?? "—" },
    { id: "desc", header: "DESCRIPTION", render: (r) => r.description ?? "—" },
  ];

  const tabs: Tab[] = useMemo(() => {
    const presentKinds = kinds.length > 0 ? kinds : Object.keys(KIND_LABELS);
    return presentKinds.map((kind) => {
      const filtered = items.filter((i) => i.kind === kind);
      return {
        id: kind,
        label: KIND_LABELS[kind] ?? kind,
        render: () => (
          <div data-testid={`${kind}-tab`}>
            <DataExplorer
              items={filtered}
              columns={columns}
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
  }, [items, kinds]);

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
