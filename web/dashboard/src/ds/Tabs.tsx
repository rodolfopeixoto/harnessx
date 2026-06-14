import { useState, type ReactNode } from "react";
import { tokens } from "./tokens";

export type Tab = {
  id: string;
  label: string;
  render: () => ReactNode;
};

type Props = {
  tabs: Tab[];
  defaultTab?: string;
};

export function Tabs({ tabs, defaultTab }: Props) {
  const [active, setActive] = useState(defaultTab || tabs[0]?.id || "");
  const current = tabs.find((t) => t.id === active) || tabs[0];
  return (
    <div data-testid="tabs">
      <div
        role="tablist"
        style={{
          display: "flex",
          gap: tokens.space(2),
          borderBottom: `1px solid ${tokens.color.border}`,
          marginBottom: tokens.space(3),
        }}
      >
        {tabs.map((t) => {
          const isActive = t.id === active;
          return (
            <button
              key={t.id}
              role="tab"
              aria-selected={isActive}
              data-testid={`tab-${t.id}`}
              onClick={() => setActive(t.id)}
              style={{
                background: "transparent",
                border: "none",
                padding: `${tokens.space(2)} ${tokens.space(3)}`,
                borderBottom: `2px solid ${isActive ? tokens.color.primary : "transparent"}`,
                color: isActive ? tokens.color.primary : tokens.color.textMuted,
                cursor: "pointer",
                fontWeight: isActive ? 600 : 400,
              }}
            >
              {t.label}
            </button>
          );
        })}
      </div>
      <div role="tabpanel" data-testid={`panel-${current?.id}`}>{current?.render()}</div>
    </div>
  );
}
