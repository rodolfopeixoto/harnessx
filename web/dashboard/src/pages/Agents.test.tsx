import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { AgentsPage } from "./Agents";

describe("AgentsPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("renders empty state when no cert", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("[]")));
    render(<MemoryRouter><AgentsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/no agent certifications yet/i)).toBeInTheDocument());
  });

  it("renders certified agents", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify([
      { agent_id: "claude", score: 100, last_certified: "2026-06-14T10:00:00Z" },
      { agent_id: "fake", score: 50, last_certified: "2026-06-14T09:00:00Z" },
    ]))));
    render(<MemoryRouter><AgentsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText("claude")).toBeInTheDocument());
    expect(screen.getByText("100/100")).toBeInTheDocument();
    expect(screen.getByText("fake")).toBeInTheDocument();
  });
});
