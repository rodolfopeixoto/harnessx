import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function AgentsPage() {
    const s = useFetched(() => api.agents(), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    if (!s.data || s.data.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no agent certifications yet (run `harness agent certify`)" } });
    return (_jsx(Panel, { title: "Agents", children: _jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "agent" }), _jsx("th", { children: "score" }), _jsx("th", { children: "last certified" })] }) }), _jsx("tbody", { children: s.data.map((a) => (_jsxs("tr", { children: [_jsx("td", { children: a.agent_id }), _jsxs("td", { children: [a.score, "/100"] }), _jsx("td", { children: a.last_certified })] }, a.agent_id))) })] }) }));
}
