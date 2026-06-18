package plancontract

import (
	"os"
	"path/filepath"
	"testing"
)

func writePlan(t *testing.T, root, id, body string) {
	t.Helper()
	d := filepath.Join(root, ".harness", "artifacts", "plans")
	_ = os.MkdirAll(d, 0o755)
	_ = os.WriteFile(filepath.Join(d, "PLAN-"+id+".md"), []byte(body), 0o644)
}

const sample = `# PLAN-X

## Intent

add /readiness endpoint

## Files in scope

- ` + "`app/main.py`" + `
- ` + "`tests/test_main.py`" + `

## Invariants

- healthz still 200

## Validation

` + "```sh" + `
harness test
` + "```" + `

## Rollback

` + "```sh" + `
git revert HEAD
` + "```" + `

## Risk tier

low
`

func TestLoadParsesSections(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "abc", sample)
	c, err := Load(dir, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if c.Intent != "add /readiness endpoint" {
		t.Errorf("intent: %q", c.Intent)
	}
	if len(c.Files) != 2 || c.Files[0] != "app/main.py" {
		t.Errorf("files: %v", c.Files)
	}
	if len(c.Invariants) != 1 {
		t.Errorf("invariants: %v", c.Invariants)
	}
	if len(c.Validation) != 1 || c.Validation[0] != "harness test" {
		t.Errorf("validation: %v", c.Validation)
	}
	if c.Rollback != "git revert HEAD" {
		t.Errorf("rollback: %q", c.Rollback)
	}
	if c.Risk != "low" {
		t.Errorf("risk: %q", c.Risk)
	}
}

func TestLoadAcceptsBareID(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "xyz", sample)
	c, err := Load(dir, "xyz")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "xyz" {
		t.Errorf("id: %q", c.ID)
	}
}

func TestLoadErrorsOnMissingIntent(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "noint", "# PLAN-noint\n\n## Files in scope\n\n- a\n")
	_, err := Load(dir, "noint")
	if err == nil {
		t.Fatal("want error")
	}
}

func TestLoadErrorsOnMissingFile(t *testing.T) {
	_, err := Load(t.TempDir(), "ghost")
	if err == nil {
		t.Fatal("want error")
	}
}

func TestInScopeUnconstrainedAllowsAny(t *testing.T) {
	c := Contract{}
	if !c.InScope("anywhere/x.py") {
		t.Error("empty scope allows any path")
	}
}

func TestInScopeMatchesGlobs(t *testing.T) {
	c := Contract{Files: []string{"app/*.py"}}
	if !c.InScope("app/main.py") {
		t.Error("main.py should match")
	}
	if c.InScope("tests/x.py") {
		t.Error("tests/x.py should not match")
	}
}

func TestResolveAbsolutePathPasses(t *testing.T) {
	got, _ := Resolve("/x", "/y/PLAN.md")
	if got != "/y/PLAN.md" {
		t.Errorf("got %s", got)
	}
}

func TestResolvePrefixesPlan(t *testing.T) {
	got, _ := Resolve("/x", "PLAN-abc.md")
	if !filepath.IsAbs(got) {
		t.Errorf("want absolute, got %s", got)
	}
}

func TestResolveAddsMissingMDSuffix(t *testing.T) {
	got, _ := Resolve("/x", "PLAN-01KV")
	if filepath.Base(got) != "PLAN-01KV.md" {
		t.Errorf("got %s", got)
	}
}

func TestResolveAddsMissingPLANPrefix(t *testing.T) {
	got, _ := Resolve("/x", "01KV")
	if filepath.Base(got) != "PLAN-01KV.md" {
		t.Errorf("got %s", got)
	}
}
