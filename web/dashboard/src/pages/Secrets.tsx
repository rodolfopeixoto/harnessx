import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function SecretsPage() {
  const s = useFetched(() => api.secretsNames(), []);
  if (s.kind !== "ready") return <PanelState state={s} />;
  const backends = Object.keys(s.data);
  return (
    <div data-testid="page-secrets" style={{ display: "grid", gap: 16 }}>
      <Panel title="Secrets (names only)">
        <p style={{ fontSize: 12, color: "#64748B" }}>
          Values are never returned over HTTP. Use <code>harness secret get &lt;name&gt; --reveal</code> on the host.
        </p>
        {backends.length === 0 ? (
          <p>no backends</p>
        ) : (
          backends.map((b) => (
            <div key={b} style={{ marginTop: 12 }}>
              <h4>{b}</h4>
              {(s.data[b] ?? []).length === 0 ? (
                <p style={{ color: "#64748B" }}>(no secrets stored)</p>
              ) : (
                <ul>{(s.data[b] ?? []).map((n) => <li key={n}><code>{n}</code></li>)}</ul>
              )}
            </div>
          ))
        )}
      </Panel>
    </div>
  );
}
