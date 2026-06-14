import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
export function PanelState({ state }) {
    switch (state.kind) {
        case "loading":
            return _jsx("p", { role: "status", children: "loading\u2026" });
        case "empty":
            return _jsx("p", { role: "status", children: state.message ?? "no data yet" });
        case "error":
            return (_jsxs("p", { role: "alert", style: { color: "#b91c1c" }, children: ["error: ", String(state.error?.message ?? state.error)] }));
        default:
            return null;
    }
}
export function Panel({ title, children }) {
    return (_jsxs("section", { style: { marginBottom: "2rem" }, "aria-label": title, children: [_jsx("h2", { style: { marginBottom: "0.5rem" }, children: title }), children] }));
}
