import { useEffect, useState } from "react";
import { Card, DataExplorer, type Column, PathCell, Badge } from "../ds";

type WorkspaceProject = {
  Slug: string;
  DisplayName: string;
  RootPath: string;
  DBPath: string;
  AddedAt: string;
  LastSeenAt?: string;
  ArchivedAt?: string;
} & Record<string, unknown>;

export function ProjectsPage() {
  const [projects, setProjects] = useState<WorkspaceProject[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void (async () => {
      try {
        const res = await fetch("/api/workspace/projects");
        if (!res.ok) {
          setError(`API ${res.status}`);
          return;
        }
        setProjects(((await res.json()) as WorkspaceProject[]) ?? []);
      } catch (err) {
        setError(String(err));
      }
    })();
  }, []);

  const columns: Column<WorkspaceProject>[] = [
    { id: "slug", header: "SLUG", render: (r) => <strong>{r.Slug}</strong> },
    { id: "name", header: "NAME", render: (r) => r.DisplayName },
    {
      id: "status",
      header: "STATUS",
      render: (r) => (
        <Badge tone={r.ArchivedAt ? "neutral" : "success"} dot>
          {r.ArchivedAt ? "archived" : "active"}
        </Badge>
      ),
    },
    { id: "path", header: "ROOT", render: (r) => <PathCell path={r.RootPath} testId="project-path-cell" /> },
    {
      id: "seen",
      header: "LAST SEEN",
      render: (r) => (r.LastSeenAt ? new Date(r.LastSeenAt).toLocaleString() : "—"),
    },
  ];

  return (
    <div data-testid="page-projects" style={{ display: "grid", gap: 16 }}>
      <Card title="Workspace projects">
        <div data-testid="project-switcher" style={{ marginBottom: 12, fontSize: 12, color: "#64748B" }}>
          {projects.length === 0 ? "no projects registered" : `${projects.length} projects · click a row to inspect`}
        </div>
        <div data-testid="projects-explorer">
          <DataExplorer
            items={projects}
            columns={columns}
            searchKeys={["Slug", "DisplayName", "RootPath"]}
            pageSize={25}
            onInspect={(row) => {
              console.info("inspect project", row.Slug);
            }}
          />
        </div>
        {error && <p style={{ color: "#B91C1C", marginTop: 12 }} role="alert">{error}</p>}
      </Card>
    </div>
  );
}
