import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Link, NavLink, Route, Routes } from "react-router-dom";
import { SessionsPage } from "./pages/Sessions";
import { SessionDetailPage } from "./pages/SessionDetail";
import { RunDetailPage } from "./pages/RunDetail";
import { SensorsPage } from "./pages/Sensors";
import { AgentsPage } from "./pages/Agents";
import { DesignPage } from "./pages/Design";
import { RoadmapPage } from "./pages/Roadmap";
import { MemoryPage } from "./pages/Memory";
import { SettingsPage } from "./pages/Settings";
const navItems = [
    { to: "/", label: "Sessions" },
    { to: "/sensors", label: "Sensors" },
    { to: "/agents", label: "Agents" },
    { to: "/design", label: "Design" },
    { to: "/roadmap", label: "Roadmap" },
    { to: "/memory", label: "Memory" },
    { to: "/settings", label: "Settings" },
];
export function App() {
    return (_jsxs("main", { style: {
            fontFamily: "-apple-system, system-ui, sans-serif",
            maxWidth: 1100,
            margin: "1.5rem auto",
            padding: "0 1rem",
        }, children: [_jsxs("header", { style: { marginBottom: "1.5rem" }, children: [_jsx(Link, { to: "/", style: { textDecoration: "none", color: "inherit" }, children: _jsx("h1", { children: "HarnessX" }) }), _jsx("nav", { "aria-label": "primary", style: { display: "flex", gap: "1rem", marginTop: "0.25rem" }, children: navItems.map((n) => (_jsx(NavLink, { to: n.to, end: n.to === "/", style: ({ isActive }) => ({
                                textDecoration: "none",
                                color: isActive ? "#4338CA" : "#374151",
                                fontWeight: isActive ? 600 : 400,
                            }), children: n.label }, n.to))) })] }), _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(SessionsPage, {}) }), _jsx(Route, { path: "/sessions/:id", element: _jsx(SessionDetailPage, {}) }), _jsx(Route, { path: "/runs/:id", element: _jsx(RunDetailPage, {}) }), _jsx(Route, { path: "/sensors", element: _jsx(SensorsPage, {}) }), _jsx(Route, { path: "/agents", element: _jsx(AgentsPage, {}) }), _jsx(Route, { path: "/design", element: _jsx(DesignPage, {}) }), _jsx(Route, { path: "/roadmap", element: _jsx(RoadmapPage, {}) }), _jsx(Route, { path: "/memory", element: _jsx(MemoryPage, {}) }), _jsx(Route, { path: "/settings", element: _jsx(SettingsPage, {}) })] })] }));
}
