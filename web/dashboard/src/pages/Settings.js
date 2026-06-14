import { jsx as _jsx, Fragment as _Fragment, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function SettingsPage() {
    const health = useFetched(() => api.health(), []);
    const profile = useFetched(() => api.profile().catch(() => null), []);
    return (_jsxs(_Fragment, { children: [_jsx(Panel, { title: "Health", children: health.kind !== "ready" ? (_jsx(PanelState, { state: health })) : (_jsx("pre", { children: JSON.stringify(health.data, null, 2) })) }), _jsx(Panel, { title: "Project profile", children: profile.kind !== "ready" ? (_jsx(PanelState, { state: profile })) : profile.data == null ? (_jsx(PanelState, { state: { kind: "empty", message: "no profile (run `harness project index`)" } })) : (_jsx("pre", { children: JSON.stringify(profile.data, null, 2) })) })] }));
}
