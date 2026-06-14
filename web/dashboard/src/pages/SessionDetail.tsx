import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function SessionDetailPage() {
  const { id = "" } = useParams();
  const s = useFetched(() => api.session(id), [id]);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const runs = s.data.runs ?? [];
  if (runs.length === 0) return <PanelState state={{ kind: "empty", message: "no runs in this session" }} />;
  return (
    <Panel title={`Session ${id.slice(0, 12)}…`}>
      <table>
        <thead>
          <tr>
            <th>run</th>
            <th>stage</th>
            <th>agent</th>
            <th>status</th>
            <th>latency</th>
            <th>cost</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((r: any) => (
            <tr key={r.id}>
              <td>
                <Link to={`/runs/${r.id}`}>{r.id.slice(0, 12)}…</Link>
              </td>
              <td>{r.stage}</td>
              <td>{r.agent ?? "—"}</td>
              <td>{r.status}</td>
              <td>{r.latency_ms ?? "—"}ms</td>
              <td>${(r.estimated_cost_usd ?? 0).toFixed?.(4)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
