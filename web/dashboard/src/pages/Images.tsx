import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function ImagesPage() {
  const s = useFetched(() => api.images(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  return (
    <div data-testid="page-images" style={{ display: "grid", gap: 16 }}>
      <Panel title="Container images">
        {s.data.length === 0 ? (
          <p>no images</p>
        ) : (
          <table>
            <thead><tr><th>repository</th><th>tag</th><th>id</th><th>created</th></tr></thead>
            <tbody>
              {s.data.map((img) => (
                <tr key={img.id}>
                  <td>{img.repository}</td>
                  <td>{img.tag}</td>
                  <td>{img.id.slice(0, 12)}</td>
                  <td>{img.created_at?.slice(0, 10) ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Panel>
    </div>
  );
}
