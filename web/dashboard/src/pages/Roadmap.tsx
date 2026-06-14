import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function RoadmapPage() {
  const s = useFetched(() => api.roadmap(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const phases = s.data?.phases ?? [];
  if (phases.length === 0)
    return <PanelState state={{ kind: "empty", message: "no roadmap yet (run `harness design-to-product`)" }} />;
  return (
    <Panel title="MVP roadmap">
      {phases.map((p: any) => (
        <article key={p.name} style={{ marginBottom: "1rem" }}>
          <h3>{p.name}</h3>
          <p>{p.goal}</p>
          {p.features && p.features.length > 0 ? (
            <ul>
              {p.features.map((f: string) => (
                <li key={f}>
                  <code>{f}</code>
                </li>
              ))}
            </ul>
          ) : (
            <p style={{ color: "#6b7280" }}>no features assigned</p>
          )}
        </article>
      ))}
    </Panel>
  );
}
