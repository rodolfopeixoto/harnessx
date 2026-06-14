import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { tokens } from "./tokens";

export type NavItem = {
  to: string;
  label: string;
  end?: boolean;
};

type Props = {
  title: string;
  nav: NavItem[];
  rightSlot?: ReactNode;
  children: ReactNode;
};

export function Shell({ title, nav, rightSlot, children }: Props) {
  return (
    <div
      data-testid="shell"
      style={{
        minHeight: "100vh",
        background: tokens.color.bg,
        fontFamily: tokens.font.family,
        color: tokens.color.text,
      }}
    >
      <header
        style={{
          padding: tokens.space(3),
          borderBottom: `1px solid ${tokens.color.border}`,
          background: tokens.color.surface,
          display: "flex",
          alignItems: "center",
          gap: tokens.space(3),
        }}
      >
        <strong>{title}</strong>
        <nav style={{ display: "flex", gap: tokens.space(2), flex: 1 }}>
          {nav.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              end={n.end}
              data-testid={`nav-${n.to.replace(/^\//, "") || "home"}`}
              style={({ isActive }) => ({
                color: isActive ? tokens.color.primary : tokens.color.textMuted,
                textDecoration: "none",
                padding: `${tokens.space(1)} ${tokens.space(2)}`,
                borderRadius: tokens.radius.sm,
                background: isActive ? `${tokens.color.primary}10` : "transparent",
                fontWeight: isActive ? 600 : 500,
                fontSize: 13,
              })}
            >
              {n.label}
            </NavLink>
          ))}
        </nav>
        {rightSlot}
      </header>
      <main style={{ padding: tokens.space(5), maxWidth: 1200, margin: "0 auto" }}>{children}</main>
    </div>
  );
}
