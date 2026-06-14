import { Link } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function SessionsPage() {
  const s = useFetched(() => api.sessions(50), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  if (!s.data || s.data.length === 0) return <PanelState state={{ kind: "empty", message: "no sessions yet" }} />;
  return (
    <Panel title="Recent sessions">
      <table>
        <thead>
          <tr>
            <th>id</th>
            <th>mode</th>
            <th>status</th>
            <th>started</th>
          </tr>
        </thead>
        <tbody>
          {s.data.map((row) => (
            <tr key={row.ID}>
              <td>
                <Link to={`/sessions/${row.ID}`}>{row.ID.slice(0, 12)}…</Link>
              </td>
              <td>{row.Mode}</td>
              <td>{row.Status}</td>
              <td>{row.StartedAt ?? ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
