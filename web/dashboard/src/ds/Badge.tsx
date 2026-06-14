import { tokens, toneColor, type Tone } from "./tokens";

type Props = {
  tone?: Tone;
  children: React.ReactNode;
  dot?: boolean;
};

export function Badge({ tone = "neutral", children, dot = false }: Props) {
  const color = toneColor(tone);
  return (
    <span
      data-testid="badge"
      data-tone={tone}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: tokens.space(1),
        padding: `${tokens.space(0.5)} ${tokens.space(2)}`,
        borderRadius: tokens.radius.sm,
        background: `${color}1A`,
        color,
        fontSize: 12,
        fontWeight: 500,
      }}
    >
      {dot && (
        <span
          aria-hidden
          style={{ width: 6, height: 6, borderRadius: "50%", background: color }}
        />
      )}
      {children}
    </span>
  );
}
