export type TerminalEntry = {
  id: string;
  command: string;
  status: "queued" | "running" | "ok" | "failed";
  stdout?: string;
  stderr?: string;
  occurredAt: string;
};

const STORAGE_KEY = "hx.terminal.entries";
const MAX_ENTRIES = 50;

function read(): TerminalEntry[] {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as TerminalEntry[];
  } catch {
    return [];
  }
}

function write(entries: TerminalEntry[]) {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(entries));
  } catch {
    /* ignore quota */
  }
}

export const TerminalReflectionService = {
  list(): TerminalEntry[] {
    return read();
  },
  record(command: string, status: TerminalEntry["status"], stdout?: string, stderr?: string): TerminalEntry {
    const entry: TerminalEntry = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      command,
      status,
      stdout,
      stderr,
      occurredAt: new Date().toISOString(),
    };
    const existing = read();
    const next = [entry, ...existing].slice(0, MAX_ENTRIES);
    write(next);
    return entry;
  },
  clear(): void {
    write([]);
  },
};
