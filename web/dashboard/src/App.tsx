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
import { Shell, type NavItem } from "./ds";

const navItems: NavItem[] = [
  { to: "/", label: "Sessions", end: true },
  { to: "/sensors", label: "Sensors" },
  { to: "/agents", label: "Agents" },
  { to: "/design", label: "Design" },
  { to: "/roadmap", label: "Roadmap" },
  { to: "/memory", label: "Memory" },
  { to: "/settings", label: "Settings" },
];

export function App() {
  return (
    <Shell title="HarnessX" nav={navItems}>
      <Routes>
        <Route path="/" element={<SessionsPage />} />
        <Route path="/sessions/:id" element={<SessionDetailPage />} />
        <Route path="/runs/:id" element={<RunDetailPage />} />
        <Route path="/sensors" element={<SensorsPage />} />
        <Route path="/agents" element={<AgentsPage />} />
        <Route path="/design" element={<DesignPage />} />
        <Route path="/roadmap" element={<RoadmapPage />} />
        <Route path="/memory" element={<MemoryPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </Shell>
  );
}
