import { useParams } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function RunDetailPage() {
  const { id = "" } = useParams();
  const s = useFetched(() => api.run(id), [id]);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const sensors = s.data.sensors ?? [];
  return (
    <Panel title={`Run ${id.slice(0, 12)}…`}>
      {sensors.length === 0 ? (
        <PanelState state={{ kind: "empty", message: "no sensor results recorded" }} />
      ) : (
        <table>
          <thead>
            <tr>
              <th>sensor</th>
              <th>status</th>
              <th>duration</th>
              <th>output</th>
            </tr>
          </thead>
          <tbody>
            {sensors.map((row: any) => (
              <tr key={row.ID}>
                <td>{row.Sensor}</td>
                <td>{row.Status}</td>
                <td>{row.DurationMs}ms</td>
                <td>{row.OutputPath ? <code>{row.OutputPath}</code> : "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </Panel>
  );
}
