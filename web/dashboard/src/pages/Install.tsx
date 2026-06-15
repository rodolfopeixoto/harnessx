import { useMemo, useState } from "react";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function InstallPage() {
  const s = useFetched(() => api.install(), []);
  const [category, setCategory] = useState("all");
  const items = s.kind === "ready" ? s.data : [];
  const categories = useMemo(() => Array.from(new Set(items.map((i) => i.category))).sort(), [items]);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const filtered = category === "all" ? items : items.filter((i) => i.category === category);
  return (
    <div data-testid="page-install" style={{ display: "grid", gap: 16 }}>
      <Panel title="Installable tools">
        <label style={{ display: "block", marginBottom: 8 }}>
          Filter:&nbsp;
          <select value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="all">all</option>
            {categories.map((c) => <option key={c} value={c}>{c}</option>)}
          </select>
        </label>
        <table>
          <thead><tr><th>name</th><th>category</th><th>binary</th><th>installed</th><th>install command</th></tr></thead>
          <tbody>
            {filtered.map((r) => (
              <tr key={r.name}>
                <td>{r.name}</td>
                <td>{r.category}</td>
                <td><code>{r.binary}</code></td>
                <td>{r.installed ? "✓" : "—"}</td>
                <td><code>harness install {r.name}</code></td>
              </tr>
            ))}
          </tbody>
        </table>
      </Panel>
    </div>
  );
}
