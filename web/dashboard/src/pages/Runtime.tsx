import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function RuntimePage() {
  const current = useFetched(() => api.runtime(), []);
  const all = useFetched(() => api.runtimes(), []);
  if (current.kind !== "ready") return <PanelState state={current} />;
  if (all.kind !== "ready") return <PanelState state={all} />;
  return (
    <div data-testid="page-runtime" style={{ display: "grid", gap: 16 }}>
      <Panel title="Current runtime">
        <table>
          <tbody>
            <tr><th align="left">runtime</th><td>{current.data.runtime ?? "(none)"}</td></tr>
            <tr><th align="left">binary</th><td>{current.data.binary}</td></tr>
            <tr><th align="left">version</th><td>{current.data.version}</td></tr>
            <tr><th align="left">source</th><td>{current.data.source}</td></tr>
          </tbody>
        </table>
      </Panel>
      <Panel title="Detected runtimes">
        <table>
          <thead><tr><th>id</th><th>binary</th><th>available</th><th>version</th><th>selected</th></tr></thead>
          <tbody>
            {all.data.map((r) => (
              <tr key={r.id}>
                <td>{r.id}</td>
                <td>{r.binary}</td>
                <td>{r.available ? "✓" : "—"}</td>
                <td>{r.version ?? "—"}</td>
                <td>{r.selected ? "★" : ""}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Panel>
    </div>
  );
}
