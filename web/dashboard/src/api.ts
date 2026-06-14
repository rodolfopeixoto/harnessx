const base = import.meta.env.VITE_API_BASE ?? "";

async function get<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(base + path, { ...init, headers: { Accept: "application/json" } });
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  return (await res.json()) as T;
}

export type Session = {
  ID: string;
  ProjectPath: string;
  Mode: string;
  Status: string;
  StartedAt?: string;
  FinishedAt?: string | null;
};

export type SensorRow = {
  id: number;
  run_id: string;
  sensor: string;
  status: "passed" | "failed" | "skipped" | string;
  duration_ms: number;
  output_path?: string | null;
  created_at: string;
};

export type Cost = {
  total_usd: number;
  by_agent?: { agent: string; cost_usd: number; input_tokens: number; output_tokens: number }[];
};

export const api = {
  health: () => get<{ ok: boolean; root: string; time: string }>("/api/health"),
  sessions: (limit = 100) => get<Session[]>(`/api/sessions?limit=${limit}`),
  session: (id: string) =>
    get<{ session_id: string; runs: any[] }>(`/api/sessions/${encodeURIComponent(id)}`),
  run: (id: string) => get<{ run_id: string; sensors: any[] }>(`/api/runs/${encodeURIComponent(id)}`),
  sensors: () => get<SensorRow[]>("/api/sensors"),
  agents: () => get<{ agent_id: string; score: number; last_certified: string }[]>("/api/agents"),
  memory: () =>
    get<{ id: string; scope: string; kind: string; content: string; confidence: number; updated_at: string }[]>(
      "/api/memory",
    ),
  cost: () => get<Cost>("/api/cost"),
  logs: (tail = 100) => get<{ lines: string[] }>(`/api/logs?tail=${tail}`),
  profile: () => get<any>("/api/profile"),
  design: () => get<any>("/api/design"),
  roadmap: () => get<any>("/api/roadmap"),
  toggles: () => get<any>("/api/toggles"),
  features: () => get<any>("/api/features"),
};
