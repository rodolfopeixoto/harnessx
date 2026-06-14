import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { RunDetailPage } from "./RunDetail";

function renderAt(id: string) {
  return render(
    <MemoryRouter initialEntries={[`/runs/${id}`]}>
      <Routes>
        <Route path="/runs/:id" element={<RunDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("RunDetailPage", () => {
  beforeEach(() => vi.unstubAllGlobals());

  it("empty sensors", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ sensors: [] }))));
    renderAt("run-xxxxxxxxxxx");
    await waitFor(() => expect(screen.getByText(/no sensor results recorded/i)).toBeInTheDocument());
  });

  it("renders sensors with output path", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
      sensors: [{ ID: 1, Sensor: "go_vet", Status: "passed", DurationMs: 200, OutputPath: ".harness/logs/go_vet.txt" }],
    }))));
    renderAt("run-xxxxxxxxxxx");
    await waitFor(() => expect(screen.getByText("go_vet")).toBeInTheDocument());
    expect(screen.getByText("passed")).toBeInTheDocument();
    expect(screen.getByText(".harness/logs/go_vet.txt")).toBeInTheDocument();
  });
});
