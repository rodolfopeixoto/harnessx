package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPyProjectTomlPEP621(t *testing.T) {
	dir := t.TempDir()
	body := `[project]
name = "todoist-api"
dependencies = [
  "fastapi>=0.110",
  "uvicorn[standard]>=0.27",
  "httpx>=0.27",
]

[project.optional-dependencies]
test = ["pytest>=8.0", "pytest-cov>=5.0"]
`
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	e, ok := readPyProjectToml(filepath.Join(dir, "pyproject.toml"))
	if !ok {
		t.Fatal("expected ok")
	}
	if e.Count < 5 {
		t.Fatalf("want >=5 deps, got %d (runtime=%d dev=%d)", e.Count, len(e.Runtime), len(e.Dev))
	}
	names := map[string]bool{}
	for _, d := range e.Runtime {
		names[d.Name] = true
	}
	for _, want := range []string{"fastapi", "uvicorn", "httpx"} {
		if !names[want] {
			t.Errorf("missing runtime dep %q", want)
		}
	}
}

func TestReadPyProjectTomlPoetry(t *testing.T) {
	dir := t.TempDir()
	body := `[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.110"
httpx = "^0.27"

[tool.poetry.group.dev.dependencies]
pytest = "^8.0"
`
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	e, ok := readPyProjectToml(filepath.Join(dir, "pyproject.toml"))
	if !ok {
		t.Fatal("expected ok")
	}
	if len(e.Runtime) != 2 {
		t.Fatalf("want 2 runtime deps (python excluded), got %d: %+v", len(e.Runtime), e.Runtime)
	}
	if len(e.Dev) != 1 {
		t.Fatalf("want 1 dev dep, got %d", len(e.Dev))
	}
}

func TestBuildDependenciesPrefersRequirementsOverPyproject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi>=0.110\nhttpx>=0.27\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\ndependencies=[\"unused\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d := BuildDependencies(dir, nil)
	py, ok := d.Ecosystems["python"]
	if !ok {
		t.Fatal("expected python ecosystem")
	}
	if py.Manifest != "requirements.txt" {
		t.Fatalf("want requirements.txt, got %s", py.Manifest)
	}
	if py.Count != 2 {
		t.Fatalf("want 2 deps from requirements.txt, got %d", py.Count)
	}
}
