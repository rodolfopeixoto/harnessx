import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ALL_ROLES, type Role } from "../auth/roles";
import { RoleProvider } from "../auth/RoleContext";
import { SessionsPage } from "../pages/Sessions";
import { SensorsPage } from "../pages/Sensors";
import { AgentsPage } from "../pages/Agents";
import { MemoryPage } from "../pages/Memory";
import { DesignPage } from "../pages/Design";
import { RoadmapPage } from "../pages/Roadmap";
import { SettingsPage } from "../pages/Settings";
import { SessionDetailPage } from "../pages/SessionDetail";
import { RunDetailPage } from "../pages/RunDetail";

type PageDef = {
  name: string;
  path: string;
  element: JSX.Element;
  awaitText: RegExp;
};

const ID_PLACEHOLDER = "demo-id";

const PAGES: PageDef[] = [
  { name: "Sessions", path: "/", element: <SessionsPage />, awaitText: /sessions yet|sessions/i },
  { name: "Sensors", path: "/sensors", element: <SensorsPage />, awaitText: /sensor runs yet|sensors/i },
  { name: "Agents", path: "/agents", element: <AgentsPage />, awaitText: /agent certifications yet|agents/i },
  { name: "Memory", path: "/memory", element: <MemoryPage />, awaitText: /project memory yet|memory/i },
  { name: "Design", path: "/design", element: <DesignPage />, awaitText: /design ingested|design/i },
  { name: "Roadmap", path: "/roadmap", element: <RoadmapPage />, awaitText: /roadmap yet|roadmap/i },
  { name: "Settings", path: "/settings", element: <SettingsPage />, awaitText: /health|profile/i },
  {
    name: "SessionDetail",
    path: `/sessions/${ID_PLACEHOLDER}`,
    element: (
      <Routes>
        <Route path="/sessions/:id" element={<SessionDetailPage />} />
      </Routes>
    ),
    awaitText: /no runs|runs/i,
  },
  {
    name: "RunDetail",
    path: `/runs/${ID_PLACEHOLDER}`,
    element: (
      <Routes>
        <Route path="/runs/:id" element={<RunDetailPage />} />
      </Routes>
    ),
    awaitText: /no sensor results|sensors/i,
  },
];

function stubFetch() {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/api/sessions/") && !url.endsWith("/sessions/")) {
        return new Response(JSON.stringify({ session_id: ID_PLACEHOLDER, runs: [] }));
      }
      if (url.includes("/api/runs/")) {
        return new Response(JSON.stringify({ run_id: ID_PLACEHOLDER, sensors: [] }));
      }
      if (url.includes("/api/sessions")) {
        return new Response("[]");
      }
      if (url.includes("/api/sensors")) {
        return new Response("[]");
      }
      if (url.includes("/api/agents")) {
        return new Response("[]");
      }
      if (url.includes("/api/memory")) {
        return new Response("[]");
      }
      if (url.includes("/api/design")) {
        return new Response(JSON.stringify({ pages: [], components: [] }));
      }
      if (url.includes("/api/roadmap")) {
        return new Response(JSON.stringify({ phases: [] }));
      }
      if (url.includes("/api/profile")) {
        return new Response(JSON.stringify({ stacks: [], generated_at: null }));
      }
      if (url.includes("/api/health")) {
        return new Response(JSON.stringify({ ok: true, root: "/tmp", time: "now" }));
      }
      return new Response("{}");
    }),
  );
}

describe("role × page grid", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
    stubFetch();
  });

  for (const page of PAGES) {
    for (const role of ALL_ROLES) {
      it(`${page.name} renders for role ${role}`, async () => {
        render(
          <RoleProvider role={role as Role}>
            <MemoryRouter initialEntries={[page.path]}>{page.element}</MemoryRouter>
          </RoleProvider>,
        );
        await waitFor(() => {
          const matches = screen.queryAllByText(page.awaitText);
          const bodyMatches = (document.body.textContent ?? "").match(page.awaitText)?.length ?? 0;
          expect(matches.length + bodyMatches).toBeGreaterThan(0);
        });
      });
    }
  }
});
