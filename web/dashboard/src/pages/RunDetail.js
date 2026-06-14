import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useParams } from "react-router-dom";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function RunDetailPage() {
    const { id = "" } = useParams();
    const s = useFetched(() => api.run(id), [id]);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    const sensors = s.data.sensors ?? [];
    return (_jsx(Panel, { title: `Run ${id.slice(0, 12)}…`, children: sensors.length === 0 ? (_jsx(PanelState, { state: { kind: "empty", message: "no sensor results recorded" } })) : (_jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "sensor" }), _jsx("th", { children: "status" }), _jsx("th", { children: "duration" }), _jsx("th", { children: "output" })] }) }), _jsx("tbody", { children: sensors.map((row) => (_jsxs("tr", { children: [_jsx("td", { children: row.Sensor }), _jsx("td", { children: row.Status }), _jsxs("td", { children: [row.DurationMs, "ms"] }), _jsx("td", { children: row.OutputPath ? _jsx("code", { children: row.OutputPath }) : "—" })] }, row.ID))) })] })) }));
}
