import type { ReactNode } from "react";
import { tokens } from "./tokens";

type Props = {
  children: ReactNode;
  onClick?: () => void;
  title?: string;
};

export function Card({ children, onClick, title }: Props) {
  return (
    <section
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      onClick={onClick}
      onKeyDown={onClick ? (e) => (e.key === "Enter" || e.key === " ") && onClick() : undefined}
      data-testid="card"
      style={{
        background: tokens.color.surface,
        border: `1px solid ${tokens.color.border}`,
        borderRadius: tokens.radius.md,
        padding: tokens.space(4),
        cursor: onClick ? "pointer" : "default",
        boxShadow: "0 1px 0 rgba(15,23,42,0.04)",
      }}
    >
      {title && <h3 style={{ margin: 0, marginBottom: tokens.space(2), fontSize: 14 }}>{title}</h3>}
      {children}
    </section>
  );
}
