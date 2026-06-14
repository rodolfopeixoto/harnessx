import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { Badge, Card, EmptyState, Tabs, InspectorPanel, DataExplorer, Shell } from "./index";

describe("Badge", () => {
  it("uses tone via data attribute", () => {
    render(<Badge tone="success" dot>online</Badge>);
    const badge = screen.getByTestId("badge");
    expect(badge.getAttribute("data-tone")).toBe("success");
    expect(badge.textContent).toContain("online");
  });
});

describe("Card", () => {
  it("becomes a button when clickable", () => {
    const onClick = vi.fn();
    render(<Card onClick={onClick}>hello</Card>);
    const card = screen.getByRole("button");
    fireEvent.click(card);
    expect(onClick).toHaveBeenCalledTimes(1);
  });
  it("triggers click on Enter key", () => {
    const onClick = vi.fn();
    render(<Card onClick={onClick}>hello</Card>);
    fireEvent.keyDown(screen.getByRole("button"), { key: "Enter" });
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});

describe("EmptyState", () => {
  it("renders title and hint", () => {
    render(<EmptyState title="No data" hint="try again" />);
    expect(screen.getByText("No data")).toBeInTheDocument();
    expect(screen.getByText("try again")).toBeInTheDocument();
  });
});

describe("Tabs", () => {
  it("switches the rendered panel", () => {
    render(
      <Tabs
        tabs={[
          { id: "a", label: "Alpha", render: () => <p>alpha-body</p> },
          { id: "b", label: "Beta", render: () => <p>beta-body</p> },
        ]}
      />,
    );
    expect(screen.getByText("alpha-body")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("tab-b"));
    expect(screen.getByText("beta-body")).toBeInTheDocument();
  });
});

describe("InspectorPanel", () => {
  it("renders title, body and closes via button + Escape", () => {
    const onClose = vi.fn();
    render(
      <InspectorPanel title="Item" subtitle="sub" body={<p>body-content</p>} onClose={onClose} />,
    );
    expect(screen.getByText("Item")).toBeInTheDocument();
    expect(screen.getByText("body-content")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("inspector-close"));
    expect(onClose).toHaveBeenCalled();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onClose.mock.calls.length).toBeGreaterThanOrEqual(2);
  });
});

describe("DataExplorer", () => {
  type Row = { name: string; status: string };
  const items: Row[] = Array.from({ length: 60 }, (_, i) => ({
    name: `item-${i}`,
    status: i % 2 ? "ok" : "fail",
  }));
  const columns = [
    { id: "name", header: "NAME", render: (r: Row) => r.name },
    { id: "status", header: "STATUS", render: (r: Row) => r.status },
  ];

  it("filters via search and updates count", () => {
    render(<DataExplorer items={items} columns={columns} searchKeys={["name"]} pageSize={20} />);
    expect(screen.getByTestId("data-explorer-count").textContent).toBe("60");
    fireEvent.change(screen.getByTestId("data-explorer-search"), { target: { value: "item-1" } });
    expect(Number(screen.getByTestId("data-explorer-count").textContent)).toBeLessThan(60);
  });

  it("paginates", () => {
    render(<DataExplorer items={items} columns={columns} searchKeys={["name"]} pageSize={20} />);
    expect(screen.getByTestId("data-explorer-page").textContent).toBe("1/3");
    fireEvent.click(screen.getByTestId("data-explorer-next"));
    expect(screen.getByTestId("data-explorer-page").textContent).toBe("2/3");
  });

  it("calls onInspect when a row is clicked", () => {
    const onInspect = vi.fn();
    render(
      <DataExplorer
        items={items.slice(0, 5)}
        columns={columns}
        searchKeys={["name"]}
        onInspect={onInspect}
      />,
    );
    fireEvent.click(screen.getAllByTestId("data-explorer-row")[0]);
    expect(onInspect).toHaveBeenCalledTimes(1);
  });

  it("renders empty state when no items", () => {
    render(<DataExplorer items={[]} columns={columns} searchKeys={["name"]} />);
    expect(screen.getByTestId("empty-state")).toBeInTheDocument();
  });
});

describe("Shell", () => {
  it("renders nav links via react-router", () => {
    render(
      <MemoryRouter initialEntries={["/sessions"]}>
        <Shell
          title="HarnessX"
          nav={[
            { to: "/", label: "Home", end: true },
            { to: "/sessions", label: "Sessions" },
          ]}
        >
          <p>content</p>
        </Shell>
      </MemoryRouter>,
    );
    expect(screen.getByText("content")).toBeInTheDocument();
    const sessionsLink = screen.getByTestId("nav-sessions");
    expect(sessionsLink).toBeInTheDocument();
  });
});
