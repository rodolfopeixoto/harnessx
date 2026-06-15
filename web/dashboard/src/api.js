const base = import.meta.env.VITE_API_BASE ?? "";
async function get(path, init) {
    const res = await fetch(base + path, { ...init, headers: { Accept: "application/json" } });
    if (!res.ok)
        throw new Error(`${path}: ${res.status}`);
    return (await res.json());
}
export const api = {
    health: () => get("/api/health"),
    runtime: () => get("/api/runtime"),
    runtimes: () => get("/api/runtimes"),
    containers: (all = false) => get(`/api/containers?all=${all}`),
    images: () => get("/api/images"),
    install: () => get("/api/install"),
    secretsNames: () => get("/api/secrets/names"),
    sessions: (limit = 100) => get(`/api/sessions?limit=${limit}`),
    session: (id) => get(`/api/sessions/${encodeURIComponent(id)}`),
    run: (id) => get(`/api/runs/${encodeURIComponent(id)}`),
    sensors: () => get("/api/sensors"),
    agents: () => get("/api/agents"),
    memory: () => get("/api/memory"),
    cost: () => get("/api/cost"),
    logs: (tail = 100) => get(`/api/logs?tail=${tail}`),
    profile: () => get("/api/profile"),
    design: () => get("/api/design"),
    roadmap: () => get("/api/roadmap"),
    toggles: () => get("/api/toggles"),
    features: () => get("/api/features"),
};
