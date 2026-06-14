import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Link } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function SessionsPage() {
    const s = useFetched(() => api.sessions(50), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    if (!s.data || s.data.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no sessions yet" } });
    return (_jsx(Panel, { title: "Recent sessions", children: _jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "id" }), _jsx("th", { children: "mode" }), _jsx("th", { children: "status" }), _jsx("th", { children: "started" })] }) }), _jsx("tbody", { children: s.data.map((row) => (_jsxs("tr", { children: [_jsx("td", { children: _jsxs(Link, { to: `/sessions/${row.ID}`, children: [row.ID.slice(0, 12), "\u2026"] }) }), _jsx("td", { children: row.Mode }), _jsx("td", { children: row.Status }), _jsx("td", { children: row.StartedAt ?? "" })] }, row.ID))) })] }) }));
}
