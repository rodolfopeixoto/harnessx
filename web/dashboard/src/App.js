import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Route, Routes } from "react-router-dom";
import { SessionsPage } from "./pages/Sessions";
import { SessionDetailPage } from "./pages/SessionDetail";
import { RunDetailPage } from "./pages/RunDetail";
import { SensorsPage } from "./pages/Sensors";
import { AgentsPage } from "./pages/Agents";
import { DesignPage } from "./pages/Design";
import { RoadmapPage } from "./pages/Roadmap";
import { MemoryPage } from "./pages/Memory";
import { SettingsPage } from "./pages/Settings";
import { Shell } from "./ds";
import { CommandPalette } from "./components/CommandPalette";
const navItems = [
    { to: "/", label: "Sessions", end: true },
    { to: "/sensors", label: "Sensors" },
    { to: "/agents", label: "Agents" },
    { to: "/design", label: "Design" },
    { to: "/roadmap", label: "Roadmap" },
    { to: "/memory", label: "Memory" },
    { to: "/settings", label: "Settings" },
];
export function App() {
    return (_jsxs(Shell, { title: "HarnessX", nav: navItems, children: [_jsx(CommandPalette, {}), _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(SessionsPage, {}) }), _jsx(Route, { path: "/sessions/:id", element: _jsx(SessionDetailPage, {}) }), _jsx(Route, { path: "/runs/:id", element: _jsx(RunDetailPage, {}) }), _jsx(Route, { path: "/sensors", element: _jsx(SensorsPage, {}) }), _jsx(Route, { path: "/agents", element: _jsx(AgentsPage, {}) }), _jsx(Route, { path: "/design", element: _jsx(DesignPage, {}) }), _jsx(Route, { path: "/roadmap", element: _jsx(RoadmapPage, {}) }), _jsx(Route, { path: "/memory", element: _jsx(MemoryPage, {}) }), _jsx(Route, { path: "/settings", element: _jsx(SettingsPage, {}) })] })] }));
}
