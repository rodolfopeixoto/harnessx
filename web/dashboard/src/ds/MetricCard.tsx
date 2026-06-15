import type { ReactNode } from "react";
import { tokens, toneColor, type Tone } from "./tokens";

type Props = {
  label: string;
  value: ReactNode;
  hint?: string;
  tone?: Tone;
  testId?: string;
};

export function MetricCard({ label, value, hint, tone = "neutral", testId }: Props) {
  const color = toneColor(tone);
  return (
    <div
      data-testid={testId}
      style={{
        background: tokens.color.surface,
        border: `1px solid ${tokens.color.border}`,
        borderRadius: tokens.radius.md,
        padding: tokens.space(4),
        minWidth: 0,
      }}
    >
      <div style={{ fontSize: 11, color: tokens.color.textMuted, textTransform: "uppercase", letterSpacing: 0.5 }}>{label}</div>
      <div style={{ fontSize: 24, fontWeight: 700, marginTop: tokens.space(1), color }}>{value}</div>
      {hint && <div style={{ fontSize: 12, color: tokens.color.textMuted, marginTop: tokens.space(1) }}>{hint}</div>}
    </div>
  );
}
