import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function DesignPage() {
  const s = useFetched(() => api.design(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const pages = s.data?.pages ?? [];
  return (
    <Panel title="Design manifest">
      <p>
        Source: <code>{s.data?.source ?? "—"}</code>. Pages: {pages.length}. Components: {s.data?.components?.length ?? 0}.
      </p>
      {pages.length === 0 ? (
        <PanelState state={{ kind: "empty", message: "no design ingested (run `harness design-to-product`)" }} />
      ) : (
        <table>
          <thead>
            <tr>
              <th>id</th>
              <th>path</th>
              <th>title</th>
              <th>components</th>
              <th>interactions</th>
            </tr>
          </thead>
          <tbody>
            {pages.map((p: any) => (
              <tr key={p.id}>
                <td>{p.id}</td>
                <td>
                  <code>{p.path}</code>
                </td>
                <td>{p.title}</td>
                <td>{(p.components ?? []).join(", ")}</td>
                <td>{(p.interactions ?? []).join(", ")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </Panel>
  );
}
