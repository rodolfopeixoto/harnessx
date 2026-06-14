import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Link, useParams } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function SessionDetailPage() {
    const { id = "" } = useParams();
    const s = useFetched(() => api.session(id), [id]);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    const runs = s.data.runs ?? [];
    if (runs.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no runs in this session" } });
    return (_jsx(Panel, { title: `Session ${id.slice(0, 12)}…`, children: _jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "run" }), _jsx("th", { children: "stage" }), _jsx("th", { children: "agent" }), _jsx("th", { children: "status" }), _jsx("th", { children: "latency" }), _jsx("th", { children: "cost" })] }) }), _jsx("tbody", { children: runs.map((r) => (_jsxs("tr", { children: [_jsx("td", { children: _jsxs(Link, { to: `/runs/${r.id}`, children: [r.id.slice(0, 12), "\u2026"] }) }), _jsx("td", { children: r.stage }), _jsx("td", { children: r.agent ?? "—" }), _jsx("td", { children: r.status }), _jsxs("td", { children: [r.latency_ms ?? "—", "ms"] }), _jsxs("td", { children: ["$", (r.estimated_cost_usd ?? 0).toFixed?.(4)] })] }, r.id))) })] }) }));
}
