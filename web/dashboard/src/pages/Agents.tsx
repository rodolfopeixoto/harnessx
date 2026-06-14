import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function AgentsPage() {
  const s = useFetched(() => api.agents(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  if (!s.data || s.data.length === 0)
    return <PanelState state={{ kind: "empty", message: "no agent certifications yet (run `harness agent certify`)" }} />;
  return (
    <Panel title="Agents">
      <table>
        <thead>
          <tr>
            <th>agent</th>
            <th>score</th>
            <th>last certified</th>
          </tr>
        </thead>
        <tbody>
          {s.data.map((a) => (
            <tr key={a.agent_id}>
              <td>{a.agent_id}</td>
              <td>{a.score}/100</td>
              <td>{a.last_certified}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
