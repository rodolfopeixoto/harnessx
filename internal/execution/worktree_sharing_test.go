package execution

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/index"
)

type writeFilesAdapter struct {
	files map[string]string
}

func (a *writeFilesAdapter) ID() string   { return "fake-writer" }
func (a *writeFilesAdapter) Name() string { return "fake writer" }
func (a *writeFilesAdapter) Capabilities() agents.Capabilities {
	return agents.Capabilities{Text: true, Diff: true}
}
func (a *writeFilesAdapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return agents.HealthcheckResult{OK: true}
}
func (a *writeFilesAdapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	for rel, body := range a.files {
		full := filepath.Join(req.WorkingDir, rel)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		_ = os.WriteFile(full, []byte(body), 0o644)
	}
	return agents.AgentResult{
		Output:   agents.AgentOutput{FinalMessage: "wrote " + flatJoin(a.files), Stdout: []byte("ok\n")},
		Usage:    agents.Usage{InputTokens: 200, OutputTokens: 50, Mode: "estimated"},
		Duration: 5 * time.Millisecond,
	}
}
func (a *writeFilesAdapter) ParseUsage(_ agents.AgentOutput) agents.Usage {
	return agents.Usage{InputTokens: 200, OutputTokens: 50, Mode: "estimated"}
}
func (a *writeFilesAdapter) ClassifyFailure(_ agents.AgentOutput, _ int, _ error) agents.FailureType {
	return agents.FailureNone
}

func flatJoin(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}

func initGitRepo(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "harness-test@example.com"},
		{"config", "user.name", "harness test"},
		{"commit", "--allow-empty", "-q", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, string(out))
		}
	}
}

func TestWorktreeIsolatesAgentWriteFromProjectRoot(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	adapter := &writeFilesAdapter{files: map[string]string{
		"app/handler.go": "package app\nfunc Sum(a, b int) int { return a + b }\n",
	}}
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{Root: root})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "add Sum",
		AgentID:  "fake-writer",
		Mode:     ModeFeature,
		Autonomy: AutonomyManual,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if res.Status == StatusAgentFailed {
		t.Fatalf("agent failed unexpectedly: %s — %s", res.Status, res.ErrorMessage)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "handler.go")); err == nil {
		t.Fatalf("worktree isolation broken: agent's file leaked to project root")
	}
	if len(res.ChangedFiles) == 0 {
		t.Fatalf("expected at least one changed file captured from worktree, got status=%s", res.Status)
	}
	if !containsString(res.ChangedFiles, "app/handler.go") {
		t.Fatalf("expected app/handler.go in changed_files, got %v", res.ChangedFiles)
	}
}

func TestWorktreeCapturesDiffAndChangedFiles(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	adapter := &writeFilesAdapter{files: map[string]string{
		"feature.txt": "hello\n",
	}}
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{Root: root})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "add hello",
		AgentID:  "fake-writer",
		Mode:     ModeFeature,
		Autonomy: AutonomyManual,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	found := false
	for _, f := range res.ChangedFiles {
		if f == "feature.txt" {
			found = true
		}
	}
	if !found {
		t.Fatalf("changed_files should include feature.txt, got %v", res.ChangedFiles)
	}
	if res.DiffPath == "" {
		t.Fatalf("expected diff path")
	}
	diff, err := os.ReadFile(res.DiffPath)
	if err != nil {
		t.Fatalf("read diff: %v", err)
	}
	if !strings.Contains(string(diff), "feature.txt") || !strings.Contains(string(diff), "+hello") {
		t.Fatalf("diff missing expected content: %s", string(diff))
	}
}

func TestWorktreeExcludesDoNotPolluteAgentDiff(t *testing.T) {
	root := t.TempDir()
	initGitRepo(t, root)

	adapter := &writeFilesAdapter{files: map[string]string{
		"src/main.py":                  "print('hi')\n",
		"__pycache__/main.cpython.pyc": "binary noise\n",
		".venv/lib/python3.12/site.py": "noise\n",
		"node_modules/foo/index.js":    "noise\n",
		"target/debug/binary":          "noise\n",
	}}
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{Root: root})
	res, err := ex.Execute(context.Background(), Request{Prompt: "x", AgentID: "fake-writer", Mode: ModeFeature, Autonomy: AutonomyManual})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, f := range res.ChangedFiles {
		for _, bad := range []string{"__pycache__/", ".venv/", "node_modules/", "target/"} {
			if strings.Contains(f, bad) {
				t.Errorf("changed_files leaked %s (file=%s)", bad, f)
			}
		}
	}
	if !containsString(res.ChangedFiles, "src/main.py") {
		t.Fatalf("expected src/main.py in changed_files, got %v", res.ChangedFiles)
	}
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
