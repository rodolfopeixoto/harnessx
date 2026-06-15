import { useState } from "react";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function ContainersPage() {
  const [all, setAll] = useState(false);
  const s = useFetched(() => api.containers(all), [all]);
  if (s.kind !== "ready") return <PanelState state={s} />;
  return (
    <div data-testid="page-containers" style={{ display: "grid", gap: 16 }}>
      <Panel title="Containers">
        <label style={{ display: "block", marginBottom: 8 }}>
          <input type="checkbox" checked={all} onChange={(e) => setAll(e.target.checked)} /> show stopped (--all)
        </label>
        {s.data.length === 0 ? (
          <p>no containers</p>
        ) : (
          <table>
            <thead><tr><th>id</th><th>name</th><th>image</th><th>status</th></tr></thead>
            <tbody>
              {s.data.map((c) => (
                <tr key={c.ID}>
                  <td>{c.ID.slice(0, 12)}</td>
                  <td>{c.Name}</td>
                  <td>{c.Image}</td>
                  <td>{c.Status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        <p style={{ marginTop: 12, fontSize: 12, color: "#64748B" }}>
          Mutations (kill/prune) require CLI: <code>harness containers kill &lt;id&gt;</code> /
          <code> harness containers prune --stopped</code>.
        </p>
      </Panel>
    </div>
  );
}
