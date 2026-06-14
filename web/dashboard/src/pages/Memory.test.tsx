import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { MemoryPage } from "./Memory";

describe("MemoryPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("renders empty state", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("[]")));
    render(<MemoryRouter><MemoryPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/no project memory yet/i)).toBeInTheDocument());
  });

  it("renders memory rows with confidence %", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify([
      { id: "m1", scope: "project", kind: "convention", content: "use rspec", confidence: 0.85, updated_at: "2026-06-14T10:00:00Z" },
    ]))));
    render(<MemoryRouter><MemoryPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText("use rspec")).toBeInTheDocument());
    expect(screen.getByText("85%")).toBeInTheDocument();
  });
});
