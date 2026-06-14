import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function SensorsPage() {
  const s = useFetched(() => api.sensors(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  if (!s.data || s.data.length === 0) return <PanelState state={{ kind: "empty", message: "no sensor runs yet" }} />;
  return (
    <Panel title="Sensor results">
      <table>
        <thead>
          <tr>
            <th>sensor</th>
            <th>status</th>
            <th>run</th>
            <th>duration</th>
            <th>at</th>
          </tr>
        </thead>
        <tbody>
          {s.data.map((row) => (
            <tr key={row.id}>
              <td>{row.sensor}</td>
              <td style={{ color: row.status === "failed" ? "#b91c1c" : "#15803d" }}>{row.status}</td>
              <td>
                <code>{row.run_id.slice(0, 12)}…</code>
              </td>
              <td>{row.duration_ms}ms</td>
              <td>{row.created_at}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
