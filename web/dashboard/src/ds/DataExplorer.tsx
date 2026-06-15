import { useMemo, useState, type ReactNode } from "react";
import { tokens } from "./tokens";
import { strings } from "./strings";
import { EmptyState } from "./EmptyState";

export type Column<T> = {
  id: string;
  header: string;
  render: (row: T) => ReactNode;
  sort?: (a: T, b: T) => number;
};

type Props<T> = {
  items: T[];
  columns: Column<T>[];
  searchKeys: (keyof T)[];
  pageSize?: number;
  onInspect?: (row: T) => void;
  emptyState?: ReactNode;
};

const DEFAULT_PAGE_SIZE = 25;

export function DataExplorer<T extends Record<string, unknown>>({
  items,
  columns,
  searchKeys,
  pageSize = DEFAULT_PAGE_SIZE,
  onInspect,
  emptyState,
}: Props<T>) {
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(0);
  const filtered = useMemo(() => {
    if (!query) return items;
    const needle = query.toLowerCase();
    return items.filter((it) =>
      searchKeys.some((key) => String(it[key] ?? "").toLowerCase().includes(needle)),
    );
  }, [items, query, searchKeys]);
  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const safePage = Math.min(page, totalPages - 1);
  const pageRows = filtered.slice(safePage * pageSize, safePage * pageSize + pageSize);

  if (items.length === 0) {
    return <>{emptyState || <EmptyState />}</>;
  }
  return (
    <div data-testid="data-explorer">
      <div
        style={{
          display: "flex",
          gap: tokens.space(2),
          marginBottom: tokens.space(3),
          alignItems: "center",
        }}
      >
        <input
          aria-label={strings.search}
          data-testid="data-explorer-search"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setPage(0);
          }}
          placeholder={strings.search}
          style={{
            flex: 1,
            padding: `${tokens.space(2)} ${tokens.space(3)}`,
            border: `1px solid ${tokens.color.border}`,
            borderRadius: tokens.radius.md,
            fontSize: 13,
          }}
        />
        <span data-testid="data-explorer-count" style={{ color: tokens.color.textMuted, fontSize: 12 }}>
          {filtered.length}
        </span>
      </div>
      <div style={{ overflowX: "auto" }}>
      <table
        style={{ width: "100%", borderCollapse: "collapse" }}
        data-testid="data-explorer-table"
      >
        <thead>
          <tr>
            {columns.map((c) => (
              <th
                key={c.id}
                style={{
                  textAlign: "left",
                  padding: tokens.space(2),
                  borderBottom: `1px solid ${tokens.color.border}`,
                  fontSize: 11,
                  textTransform: "uppercase",
                  color: tokens.color.textMuted,
                  letterSpacing: 0.5,
                }}
              >
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {pageRows.map((row, idx) => (
            <tr
              key={idx}
              data-testid="data-explorer-row"
              onClick={onInspect ? () => onInspect(row) : undefined}
              style={{ cursor: onInspect ? "pointer" : "default" }}
            >
              {columns.map((c) => (
                <td
                  key={c.id}
                  style={{
                    padding: tokens.space(2),
                    borderBottom: `1px solid ${tokens.color.border}`,
                    fontSize: 13,
                  }}
                >
                  {c.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      </div>
      {totalPages > 1 && (
        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: tokens.space(2),
            marginTop: tokens.space(3),
          }}
        >
          <button
            data-testid="data-explorer-prev"
            disabled={safePage === 0}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
          >
            ‹
          </button>
          <span data-testid="data-explorer-page">{safePage + 1}/{totalPages}</span>
          <button
            data-testid="data-explorer-next"
            disabled={safePage >= totalPages - 1}
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
          >
            ›
          </button>
        </div>
      )}
    </div>
  );
}
