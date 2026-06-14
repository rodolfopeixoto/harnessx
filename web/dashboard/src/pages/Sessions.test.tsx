import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SessionsPage } from "./Sessions";

function mockFetch(payload: unknown, status = 200) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      new Response(JSON.stringify(payload), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
    ),
  );
}

describe("SessionsPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("shows loading state then empty when API returns []", async () => {
    mockFetch([]);
    render(<MemoryRouter><SessionsPage /></MemoryRouter>);
    expect(screen.getByRole("status")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/no sessions yet/i)).toBeInTheDocument());
  });

  it("renders a row per session with link", async () => {
    mockFetch([
      { ID: "01ABCDEFGHJKMNPQRSTVWXYZ01", Mode: "feature", Status: "succeeded", StartedAt: "2026-06-14T10:00:00Z" },
    ]);
    render(<MemoryRouter><SessionsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/feature/)).toBeInTheDocument());
    expect(screen.getByText(/succeeded/)).toBeInTheDocument();
    expect(screen.getByRole("link")).toHaveAttribute("href", "/sessions/01ABCDEFGHJKMNPQRSTVWXYZ01");
  });

  it("renders error state when fetch fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("", { status: 500 })));
    render(<MemoryRouter><SessionsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
  });
});
