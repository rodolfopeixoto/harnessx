import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function MemoryPage() {
    const s = useFetched(() => api.memory(), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    if (!s.data || s.data.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no project memory yet" } });
    return (_jsx(Panel, { title: "Project memory", children: _jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "scope" }), _jsx("th", { children: "kind" }), _jsx("th", { children: "content" }), _jsx("th", { children: "confidence" }), _jsx("th", { children: "updated" })] }) }), _jsx("tbody", { children: s.data.map((m) => (_jsxs("tr", { children: [_jsx("td", { children: m.scope }), _jsx("td", { children: m.kind }), _jsx("td", { children: m.content }), _jsxs("td", { children: [(m.confidence * 100).toFixed(0), "%"] }), _jsx("td", { children: m.updated_at })] }, m.id))) })] }) }));
}
