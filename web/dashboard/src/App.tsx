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
  return (
    <main
      style={{
        fontFamily: "-apple-system, system-ui, sans-serif",
        maxWidth: 1100,
        margin: "1.5rem auto",
        padding: "0 1rem",
      }}
    >
      <header style={{ marginBottom: "1.5rem" }}>
        <Link to="/" style={{ textDecoration: "none", color: "inherit" }}>
          <h1>HarnessX</h1>
        </Link>
        <nav aria-label="primary" style={{ display: "flex", gap: "1rem", marginTop: "0.25rem" }}>
          {navItems.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              end={n.to === "/"}
              style={({ isActive }) => ({
                textDecoration: "none",
                color: isActive ? "#4338CA" : "#374151",
                fontWeight: isActive ? 600 : 400,
              })}
            >
              {n.label}
            </NavLink>
          ))}
        </nav>
      </header>
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
    </main>
  );
}
