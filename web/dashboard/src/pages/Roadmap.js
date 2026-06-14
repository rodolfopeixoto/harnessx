import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function RoadmapPage() {
    const s = useFetched(() => api.roadmap(), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    const phases = s.data?.phases ?? [];
    if (phases.length === 0)
        return _jsx(PanelState, { state: { kind: "empty", message: "no roadmap yet (run `harness design-to-product`)" } });
    return (_jsx(Panel, { title: "MVP roadmap", children: phases.map((p) => (_jsxs("article", { style: { marginBottom: "1rem" }, children: [_jsx("h3", { children: p.name }), _jsx("p", { children: p.goal }), p.features && p.features.length > 0 ? (_jsx("ul", { children: p.features.map((f) => (_jsx("li", { children: _jsx("code", { children: f }) }, f))) })) : (_jsx("p", { style: { color: "#6b7280" }, children: "no features assigned" }))] }, p.name))) }));
}
