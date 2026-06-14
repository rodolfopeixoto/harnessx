import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { CommandPalette } from "./CommandPalette";

function withFetch(payload: unknown, status = 200) {
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

describe("CommandPalette", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("opens on Cmd+K and closes on Escape", () => {
    render(
      <MemoryRouter>
        <CommandPalette />
      </MemoryRouter>,
    );
    fireEvent.keyDown(window, { key: "k", metaKey: true });
    expect(screen.getByTestId("palette")).toBeInTheDocument();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(screen.queryByTestId("palette")).not.toBeInTheDocument();
  });

  it("fetches results from /api/palette", async () => {
    withFetch([
      { source: "commands", kind: "command", title: "Open settings", router_path: "/settings", score: 80 },
    ]);
    render(
      <MemoryRouter>
        <CommandPalette />
      </MemoryRouter>,
    );
    fireEvent.keyDown(window, { key: "k", metaKey: true });
    fireEvent.change(screen.getByTestId("palette-input"), { target: { value: "settings" } });
    await waitFor(() => expect(screen.getByTestId("palette-hit-0")).toBeInTheDocument());
    expect(screen.getByText("Open settings")).toBeInTheDocument();
  });

  it("shows empty state when there are no hits", async () => {
    withFetch([]);
    render(
      <MemoryRouter>
        <CommandPalette />
      </MemoryRouter>,
    );
    fireEvent.keyDown(window, { key: "k", metaKey: true });
    fireEvent.change(screen.getByTestId("palette-input"), { target: { value: "xyz" } });
    await waitFor(() => expect(screen.getByTestId("palette-empty")).toBeInTheDocument());
  });
});
