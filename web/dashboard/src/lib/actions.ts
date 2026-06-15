import { TerminalReflectionService } from "./terminal";

export type ActionEvent = {
  id: string;
  kind: string;
  subject: string;
  status: "ok" | "failed";
  command?: string;
  occurredAt: string;
};

const STORAGE_KEY = "hx.actions.events";
const MAX_EVENTS = 100;

function read(): ActionEvent[] {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as ActionEvent[];
  } catch {
    return [];
  }
}

function write(events: ActionEvent[]) {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(events));
  } catch {
    /* ignore */
  }
}

export const ActionService = {
  list(): ActionEvent[] {
    return read();
  },
  record(kind: string, subject: string, command?: string, status: "ok" | "failed" = "ok"): ActionEvent {
    const event: ActionEvent = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      kind,
      subject,
      status,
      command,
      occurredAt: new Date().toISOString(),
    };
    const existing = read();
    const next = [event, ...existing].slice(0, MAX_EVENTS);
    write(next);
    if (command) {
      TerminalReflectionService.record(command, status === "ok" ? "ok" : "failed");
    }
    return event;
  },
  clear(): void {
    write([]);
  },
};
