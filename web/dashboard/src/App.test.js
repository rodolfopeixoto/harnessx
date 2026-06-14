import { jsx as _jsx } from "react/jsx-runtime";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { App } from "./App";
beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn(async (input) => {
        const url = String(input);
        const body = url.includes("/api/sessions") ? "[]" :
            url.includes("/api/health") ? '{"ok":true,"root":"/x","time":"now"}' :
                "{}";
        return new Response(body, { status: 200, headers: { "Content-Type": "application/json" } });
    }));
});
describe("App", () => {
    it("renders heading + nav", () => {
        render(_jsx(MemoryRouter, { children: _jsx(App, {}) }));
        expect(screen.getByRole("heading", { name: /HarnessX/i })).toBeInTheDocument();
        expect(screen.getByRole("navigation")).toBeInTheDocument();
        expect(screen.getByRole("link", { name: /Sensors/i })).toHaveAttribute("href", "/sensors");
    });
});
