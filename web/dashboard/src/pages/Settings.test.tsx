import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { SettingsPage } from "./Settings";

describe("SettingsPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("renders health + profile", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/health")) {
          return new Response(JSON.stringify({ ok: true, root: "/tmp", time: "now" }));
        }
        if (url.includes("/api/profile")) {
          return new Response(JSON.stringify({ stacks: [{ name: "go" }] }));
        }
        return new Response("{}");
      }),
    );
    render(<MemoryRouter><SettingsPage /></MemoryRouter>);
    await waitFor(() => expect(screen.getByText(/Health/i)).toBeInTheDocument());
    expect(screen.getByText(/Project profile/i)).toBeInTheDocument();
    expect(screen.getAllByText(/"ok": true/i).length).toBeGreaterThan(0);
  });
});
