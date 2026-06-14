import type { ReactNode } from "react";

type State =
  | { kind: "loading" }
  | { kind: "empty"; message?: string }
  | { kind: "error"; error: unknown }
  | { kind: "ready" };

export function PanelState({ state }: { state: State }) {
  switch (state.kind) {
    case "loading":
      return <p role="status">loading…</p>;
    case "empty":
      return <p role="status">{state.message ?? "no data yet"}</p>;
    case "error":
      return (
        <p role="alert" style={{ color: "#b91c1c" }}>
          error: {String((state.error as Error)?.message ?? state.error)}
        </p>
      );
    default:
      return null;
  }
}

export function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section style={{ marginBottom: "2rem" }} aria-label={title}>
      <h2 style={{ marginBottom: "0.5rem" }}>{title}</h2>
      {children}
    </section>
  );
}
