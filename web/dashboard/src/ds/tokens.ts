export const tokens = {
  color: {
    bg: "#FAFAFC",
    surface: "#FFFFFF",
    border: "#E4E4E7",
    text: "#0F172A",
    textMuted: "#64748B",
    primary: "#4338CA",
    primaryFg: "#FFFFFF",
    success: "#15803D",
    warning: "#B45309",
    danger: "#B91C1C",
    info: "#1D4ED8",
  },
  radius: {
    sm: "6px",
    md: "10px",
    lg: "14px",
  },
  space: (n: number) => `${n * 4}px`,
  font: {
    family: "-apple-system, system-ui, sans-serif",
    mono: "ui-monospace, SFMono-Regular, Menlo, monospace",
  },
  z: {
    drawer: 40,
    toast: 50,
    inspector: 60,
  },
} as const;

export type Tone = "primary" | "success" | "warning" | "danger" | "info" | "neutral";

export const toneColor = (tone: Tone) => {
  switch (tone) {
    case "primary":
      return tokens.color.primary;
    case "success":
      return tokens.color.success;
    case "warning":
      return tokens.color.warning;
    case "danger":
      return tokens.color.danger;
    case "info":
      return tokens.color.info;
    default:
      return tokens.color.textMuted;
  }
};
