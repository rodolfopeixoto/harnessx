import type { ReactNode } from "react";
import { tokens } from "./tokens";
import { strings } from "./strings";

type Props = {
  title?: string;
  hint?: ReactNode;
  action?: ReactNode;
};

export function EmptyState({ title = strings.emptyTitle, hint = strings.emptyHint, action }: Props) {
  return (
    <div
      role="status"
      data-testid="empty-state"
      style={{
        padding: tokens.space(8),
        textAlign: "center",
        color: tokens.color.textMuted,
        border: `1px dashed ${tokens.color.border}`,
        borderRadius: tokens.radius.md,
      }}
    >
      <h3 style={{ margin: 0, color: tokens.color.text, fontSize: 16 }}>{title}</h3>
      {hint && <p style={{ margin: `${tokens.space(2)} 0 0` }}>{hint}</p>}
      {action && <div style={{ marginTop: tokens.space(4) }}>{action}</div>}
    </div>
  );
}
