import type { ReactNode } from "react";
import { useEffect } from "react";
import { tokens } from "./tokens";
import { strings } from "./strings";
import { Tabs, type Tab } from "./Tabs";

type Props = {
  title: string;
  subtitle?: string;
  tabs?: Tab[];
  body?: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
};

export function InspectorPanel({ title, subtitle, tabs, body, footer, onClose }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);
  return (
    <aside
      role="dialog"
      aria-modal="true"
      aria-label={title}
      data-testid="inspector"
      style={{
        position: "fixed",
        top: 0,
        right: 0,
        bottom: 0,
        width: "min(960px, 52vw)",
        background: tokens.color.surface,
        boxShadow: "-8px 0 24px rgba(15,23,42,0.12)",
        display: "flex",
        flexDirection: "column",
        zIndex: tokens.z.inspector,
      }}
    >
      <header
        style={{
          padding: tokens.space(4),
          borderBottom: `1px solid ${tokens.color.border}`,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: tokens.space(2),
          position: "sticky",
          top: 0,
          background: tokens.color.surface,
        }}
      >
        <div>
          <h2 style={{ margin: 0, fontSize: 16 }}>{title}</h2>
          {subtitle && <p style={{ margin: 0, fontSize: 12, color: tokens.color.textMuted }}>{subtitle}</p>}
        </div>
        <button
          aria-label={strings.close}
          data-testid="inspector-close"
          onClick={onClose}
          style={{
            background: "transparent",
            border: `1px solid ${tokens.color.border}`,
            borderRadius: tokens.radius.sm,
            padding: `${tokens.space(1)} ${tokens.space(2)}`,
            cursor: "pointer",
            color: tokens.color.text,
          }}
        >
          {strings.close}
        </button>
      </header>
      <div style={{ flex: 1, overflow: "auto", padding: tokens.space(4) }}>
        {tabs && tabs.length > 0 ? <Tabs tabs={tabs} /> : body}
      </div>
      {footer && (
        <footer
          style={{
            padding: tokens.space(4),
            borderTop: `1px solid ${tokens.color.border}`,
            position: "sticky",
            bottom: 0,
            background: tokens.color.surface,
            display: "flex",
            flexWrap: "wrap",
            gap: tokens.space(2),
            justifyContent: "flex-end",
          }}
        >
          {footer}
        </footer>
      )}
    </aside>
  );
}
