import { useState } from "react";
import { tokens } from "./tokens";

type Props = {
  path: string;
  testId?: string;
};

const PREFIX_KEEP = 18;
const SUFFIX_KEEP = 22;

function truncate(path: string): string {
  if (path.length <= PREFIX_KEEP + SUFFIX_KEEP + 1) return path;
  return `${path.slice(0, PREFIX_KEEP)}…${path.slice(-SUFFIX_KEEP)}`;
}

export function PathCell({ path, testId }: Props) {
  const [copied, setCopied] = useState(false);
  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(path);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* ignore */
    }
  };
  return (
    <span
      data-testid={testId}
      title={path}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: tokens.space(2),
        fontFamily: tokens.font.mono,
        fontSize: 12,
      }}
    >
      <span style={{ color: tokens.color.text }}>{truncate(path)}</span>
      <button
        type="button"
        onClick={onCopy}
        data-testid={`${testId}-copy`}
        style={{
          background: "transparent",
          border: `1px solid ${tokens.color.border}`,
          borderRadius: tokens.radius.sm,
          padding: `0 ${tokens.space(1.5)}`,
          fontSize: 11,
          cursor: "pointer",
          color: tokens.color.textMuted,
        }}
      >
        {copied ? "copied" : "copy"}
      </button>
    </span>
  );
}
