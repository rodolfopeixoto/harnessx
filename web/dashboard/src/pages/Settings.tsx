import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";

export function SettingsPage() {
  const health = useFetched(() => api.health(), []);
  const profile = useFetched(() => api.profile().catch(() => null), []);
  return (
    <>
      <Panel title="Health">
        {health.kind !== "ready" ? (
          <PanelState state={health} />
        ) : (
          <pre>{JSON.stringify(health.data, null, 2)}</pre>
        )}
      </Panel>
      <Panel title="Project profile">
        {profile.kind !== "ready" ? (
          <PanelState state={profile} />
        ) : profile.data == null ? (
          <PanelState state={{ kind: "empty", message: "no profile (run `harness project index`)" }} />
        ) : (
          <pre>{JSON.stringify(profile.data, null, 2)}</pre>
        )}
      </Panel>
    </>
  );
}
