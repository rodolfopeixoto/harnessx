import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SensorsPage } from "./Sensors";

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

describe("SensorsPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("renders empty state", async () => {
    mockFetch([]);
    render(<MemoryRouter><SensorsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/no sensor runs yet/i)).toBeInTheDocument());
  });

  it("colours failed sensors red and passed green", async () => {
    mockFetch([
      { id: 1, run_id: "r-xxxxxxxxxxxx", sensor: "go_vet", status: "passed", duration_ms: 100, created_at: "2026-06-14T10:00:00Z" },
      { id: 2, run_id: "r-yyyyyyyyyyyy", sensor: "go_test", status: "failed", duration_ms: 500, created_at: "2026-06-14T10:01:00Z" },
    ]);
    render(<MemoryRouter><SensorsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText("go_vet")).toBeInTheDocument());
    expect(screen.getByText("go_test")).toBeInTheDocument();
  });
});
