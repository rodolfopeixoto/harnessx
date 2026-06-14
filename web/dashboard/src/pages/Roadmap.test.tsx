import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { RoadmapPage } from "./Roadmap";

describe("RoadmapPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("empty when no phases", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ phases: [] }))));
    render(<MemoryRouter><RoadmapPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/no roadmap yet/i)).toBeInTheDocument());
  });

  it("renders each phase with its features", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      phases: [
        { name: "MVP 0", goal: "React parity", features: ["feature.home"] },
        { name: "MVP 1", goal: "Core flows", features: [] },
      ],
    }))));
    render(<MemoryRouter><RoadmapPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText("MVP 0")).toBeInTheDocument());
    expect(screen.getByText("MVP 1")).toBeInTheDocument();
    expect(screen.getByText("feature.home")).toBeInTheDocument();
  });
});
