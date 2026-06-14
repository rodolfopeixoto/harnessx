import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function SensorsPage() {
    const s = useFetched(() => api.sensors(), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    if (!s.data || s.data.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no sensor runs yet" } });
    return (_jsx(Panel, { title: "Sensor results", children: _jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "sensor" }), _jsx("th", { children: "status" }), _jsx("th", { children: "run" }), _jsx("th", { children: "duration" }), _jsx("th", { children: "at" })] }) }), _jsx("tbody", { children: s.data.map((row) => (_jsxs("tr", { children: [_jsx("td", { children: row.sensor }), _jsx("td", { style: { color: row.status === "failed" ? "#b91c1c" : "#15803d" }, children: row.status }), _jsx("td", { children: _jsxs("code", { children: [row.run_id.slice(0, 12), "\u2026"] }) }), _jsxs("td", { children: [row.duration_ms, "ms"] }), _jsx("td", { children: row.created_at })] }, row.id))) })] }) }));
}
