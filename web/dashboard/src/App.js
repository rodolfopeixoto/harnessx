import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { Route, Routes } from "react-router-dom";
import { HomePage } from "./pages/Home";
import { SessionsPage } from "./pages/Sessions";
import { SessionDetailPage } from "./pages/SessionDetail";
import { RunDetailPage } from "./pages/RunDetail";
import { SensorsPage } from "./pages/Sensors";
import { AgentsPage } from "./pages/Agents";
import { DesignPage } from "./pages/Design";
import { RoadmapPage } from "./pages/Roadmap";
import { MemoryPage } from "./pages/Memory";
import { SettingsPage } from "./pages/Settings";
import { ProjectsPage } from "./pages/Projects";
import { CommandPage } from "./pages/Command";
import { PlanPage } from "./pages/Plan";
import { ActiveRunPage } from "./pages/ActiveRun";
import { CatalogPage } from "./pages/Catalog";
import { ContextPage } from "./pages/Context";
import { ResourcesPage } from "./pages/Resources";
import { ReportsPage } from "./pages/Reports";
import { CleanupPage } from "./pages/Cleanup";
import { StakeholderPage } from "./pages/Stakeholder";
import { OnboardingPage } from "./pages/Onboarding";
import { RuntimePage } from "./pages/Runtime";
import { ContainersPage } from "./pages/Containers";
import { ImagesPage } from "./pages/Images";
import { InstallPage } from "./pages/Install";
import { SecretsPage } from "./pages/Secrets";
import { BackupPage } from "./pages/Backup";
import { Shell } from "./ds";
import { CommandPalette } from "./components/CommandPalette";
const navItems = [
    { to: "/", label: "Home", end: true },
    { to: "/projects", label: "Projects" },
    { to: "/command", label: "Command" },
    { to: "/run", label: "Run" },
    { to: "/plan", label: "Plan" },
    { to: "/design", label: "Design" },
    { to: "/roadmap", label: "Roadmap" },
    { to: "/agents", label: "Agents" },
    { to: "/catalog", label: "Capabilities" },
    { to: "/sensors", label: "Sensors" },
    { to: "/context", label: "Context" },
    { to: "/memory", label: "Memory" },
    { to: "/resources", label: "Resources" },
    { to: "/runtime", label: "Runtime" },
    { to: "/containers", label: "Containers" },
    { to: "/images", label: "Images" },
    { to: "/install", label: "Install" },
    { to: "/secrets", label: "Secrets" },
    { to: "/backup", label: "Backup" },
    { to: "/cleanup", label: "Cleanup" },
    { to: "/reports", label: "Reports" },
    { to: "/stakeholder", label: "Stakeholder" },
    { to: "/settings", label: "Settings" },
];
export function App() {
    return (_jsxs(Shell, { title: "HarnessX", nav: navItems, children: [_jsx(CommandPalette, {}), _jsxs(Routes, { children: [_jsx(Route, { path: "/", element: _jsx(HomePage, {}) }), _jsx(Route, { path: "/sessions", element: _jsx(SessionsPage, {}) }), _jsx(Route, { path: "/sessions/:id", element: _jsx(SessionDetailPage, {}) }), _jsx(Route, { path: "/runs/:id", element: _jsx(RunDetailPage, {}) }), _jsx(Route, { path: "/projects", element: _jsx(ProjectsPage, {}) }), _jsx(Route, { path: "/command", element: _jsx(CommandPage, {}) }), _jsx(Route, { path: "/plan", element: _jsx(PlanPage, {}) }), _jsx(Route, { path: "/run", element: _jsx(ActiveRunPage, {}) }), _jsx(Route, { path: "/design", element: _jsx(DesignPage, {}) }), _jsx(Route, { path: "/roadmap", element: _jsx(RoadmapPage, {}) }), _jsx(Route, { path: "/agents", element: _jsx(AgentsPage, {}) }), _jsx(Route, { path: "/catalog", element: _jsx(CatalogPage, {}) }), _jsx(Route, { path: "/sensors", element: _jsx(SensorsPage, {}) }), _jsx(Route, { path: "/context", element: _jsx(ContextPage, {}) }), _jsx(Route, { path: "/memory", element: _jsx(MemoryPage, {}) }), _jsx(Route, { path: "/resources", element: _jsx(ResourcesPage, {}) }), _jsx(Route, { path: "/runtime", element: _jsx(RuntimePage, {}) }), _jsx(Route, { path: "/containers", element: _jsx(ContainersPage, {}) }), _jsx(Route, { path: "/images", element: _jsx(ImagesPage, {}) }), _jsx(Route, { path: "/install", element: _jsx(InstallPage, {}) }), _jsx(Route, { path: "/secrets", element: _jsx(SecretsPage, {}) }), _jsx(Route, { path: "/backup", element: _jsx(BackupPage, {}) }), _jsx(Route, { path: "/cleanup", element: _jsx(CleanupPage, {}) }), _jsx(Route, { path: "/reports", element: _jsx(ReportsPage, {}) }), _jsx(Route, { path: "/stakeholder", element: _jsx(StakeholderPage, {}) }), _jsx(Route, { path: "/onboarding", element: _jsx(OnboardingPage, {}) }), _jsx(Route, { path: "/settings", element: _jsx(SettingsPage, {}) })] })] }));
}
