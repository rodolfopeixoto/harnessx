import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { DesignPage } from "./Design";

describe("DesignPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("renders empty when no pages", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ pages: [], components: [] }))));
    render(<MemoryRouter><DesignPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/no design ingested/i)).toBeInTheDocument());
  });

  it("renders pages table with counts", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      source: "design.zip",
      pages: [{ id: "home", path: "/", title: "Home", components: ["UiHero"], interactions: ["click"] }],
      components: [{ name: "UiHero" }],
    }))));
    render(<MemoryRouter><DesignPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText("home")).toBeInTheDocument());
    expect(screen.getByText("Home")).toBeInTheDocument();
    expect(screen.getByText("UiHero")).toBeInTheDocument();
  });
});
