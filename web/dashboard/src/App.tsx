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
import { Shell, type NavItem } from "./ds";
import { CommandPalette } from "./components/CommandPalette";

const navItems: NavItem[] = [
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
  { to: "/cleanup", label: "Cleanup" },
  { to: "/reports", label: "Reports" },
  { to: "/stakeholder", label: "Stakeholder" },
  { to: "/settings", label: "Settings" },
];

export function App() {
  return (
    <Shell title="HarnessX" nav={navItems}>
      <CommandPalette />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/sessions" element={<SessionsPage />} />
        <Route path="/sessions/:id" element={<SessionDetailPage />} />
        <Route path="/runs/:id" element={<RunDetailPage />} />
        <Route path="/projects" element={<ProjectsPage />} />
        <Route path="/command" element={<CommandPage />} />
        <Route path="/plan" element={<PlanPage />} />
        <Route path="/run" element={<ActiveRunPage />} />
        <Route path="/design" element={<DesignPage />} />
        <Route path="/roadmap" element={<RoadmapPage />} />
        <Route path="/agents" element={<AgentsPage />} />
        <Route path="/catalog" element={<CatalogPage />} />
        <Route path="/sensors" element={<SensorsPage />} />
        <Route path="/context" element={<ContextPage />} />
        <Route path="/memory" element={<MemoryPage />} />
        <Route path="/resources" element={<ResourcesPage />} />
        <Route path="/cleanup" element={<CleanupPage />} />
        <Route path="/reports" element={<ReportsPage />} />
        <Route path="/stakeholder" element={<StakeholderPage />} />
        <Route path="/onboarding" element={<OnboardingPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </Shell>
  );
}
