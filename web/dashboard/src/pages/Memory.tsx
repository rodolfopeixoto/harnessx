import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function MemoryPage() {
  const s = useFetched(() => api.memory(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  if (!s.data || s.data.length === 0)
    return <PanelState state={{ kind: "empty", message: "no project memory yet" }} />;
  return (
    <Panel title="Project memory">
      <table>
        <thead>
          <tr>
            <th>scope</th>
            <th>kind</th>
            <th>content</th>
            <th>confidence</th>
            <th>updated</th>
          </tr>
        </thead>
        <tbody>
          {s.data.map((m) => (
            <tr key={m.id}>
              <td>{m.scope}</td>
              <td>{m.kind}</td>
              <td>{m.content}</td>
              <td>{(m.confidence * 100).toFixed(0)}%</td>
              <td>{m.updated_at}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
