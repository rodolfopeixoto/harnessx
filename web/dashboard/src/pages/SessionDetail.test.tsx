import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SessionDetailPage } from "./SessionDetail";

function renderAt(id: string) {
  return render(
    <MemoryRouter initialEntries={[`/sessions/${id}`]}>
      <Routes>
        <Route path="/sessions/:id" element={<SessionDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("SessionDetailPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("empty when no runs", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ runs: [] }))));
    renderAt("01ABCDEFGHJKMNPQRSTVWXYZ01");
    await waitFor(() => expect(screen.getByText(/no runs in this session/i)).toBeInTheDocument());
  });

  it("renders runs table with link to run detail", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      runs: [{ id: "run-xxxxxxxxxxx", stage: "execute", agent: "claude", status: "succeeded", latency_ms: 1200, estimated_cost_usd: 0.0123 }],
    }))));
    renderAt("01ABCDEFGHJKMNPQRSTVWXYZ01");
    await waitFor(() => expect(screen.getByText("execute")).toBeInTheDocument());
    expect(screen.getByText("claude")).toBeInTheDocument();
    expect(screen.getByText("1200ms")).toBeInTheDocument();
    expect(screen.getByRole("link")).toHaveAttribute("href", "/runs/run-xxxxxxxxxxx");
  });
});
