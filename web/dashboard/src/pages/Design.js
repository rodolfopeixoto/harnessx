import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { api } from "../api";
import { Panel, PanelState } from "../components/Panel";
import { useFetched } from "../hooks";
export function DesignPage() {
    const s = useFetched(() => api.design(), []);
    if (s.kind !== "ready")
        return _jsx(PanelState, { state: s });
    const pages = s.data?.pages ?? [];
    return (_jsxs(Panel, { title: "Design manifest", children: [_jsxs("p", { children: ["Source: ", _jsx("code", { children: s.data?.source ?? "—" }), ". Pages: ", pages.length, ". Components: ", s.data?.components?.length ?? 0, "."] }), pages.length === 0 ? (_jsx(PanelState, { state: { kind: "empty", message: "no design ingested (run `harness design-to-product`)" } })) : (_jsxs("table", { children: [_jsx("thead", { children: _jsxs("tr", { children: [_jsx("th", { children: "id" }), _jsx("th", { children: "path" }), _jsx("th", { children: "title" }), _jsx("th", { children: "components" }), _jsx("th", { children: "interactions" })] }) }), _jsx("tbody", { children: pages.map((p) => (_jsxs("tr", { children: [_jsx("td", { children: p.id }), _jsx("td", { children: _jsx("code", { children: p.path }) }), _jsx("td", { children: p.title }), _jsx("td", { children: (p.components ?? []).join(", ") }), _jsx("td", { children: (p.interactions ?? []).join(", ") })] }, p.id))) })] }))] }));
}
