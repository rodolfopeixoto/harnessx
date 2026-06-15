// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents/yaml"
	"github.com/ropeixoto/harnessx/internal/index"
)

func buildFakeAgent(t *testing.T) string {
	t.Helper()
	repoRoot := findRepoRoot(t)
	bin := filepath.Join(t.TempDir(), "harness-fake-agent")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/fake-agent")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GOROOT=")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake-agent: %v: %s", err, out)
	}
	return bin
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for d := wd; d != "/"; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
	}
	t.Fatal("repo root not found")
	return ""
}

func fakeSpec(binPath string) yaml.Spec {
	var s yaml.Spec
	s.ID = "fake-real"
	s.Name = "fake"
	s.Type = "cli"
	s.Command.Binary = binPath
	s.Command.Check = binPath + " --help"
	s.Capabilities.Text = true
	s.Capabilities.Files = true
	s.Capabilities.JSONOutput = true
	s.Execution.PromptMode = "stdin"
	s.Execution.WorkingDirectory = "project"
	s.Execution.TimeoutSeconds = 60
	s.Output.Format = "json"
	s.Output.FinalMessageJSONPath = "$.result"
	s.Output.UsageJSONPath = "$.usage"
	return s
}

func TestExecute_FakeAgentProducesDiff(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	bin := buildFakeAgent(t)
	adapter := yaml.New(fakeSpec(bin))
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "create greet.md with content: hello",
		Mode:     ModeFeature,
		Apply:    false,
		Autonomy: AutonomySafeExecute,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Status != StatusWaitingApproval {
		t.Fatalf("expected waiting_approval, got %s (err=%s)", res.Status, res.ErrorMessage)
	}
	if len(res.ChangedFiles) != 1 || res.ChangedFiles[0] != "greet.md" {
		t.Fatalf("unexpected changed files: %v", res.ChangedFiles)
	}
	if res.DiffPath == "" {
		t.Fatal("diff path empty")
	}
	if _, err := os.Stat(res.DiffPath); err != nil {
		t.Fatalf("diff.patch missing: %v", err)
	}
	if _, err := os.Stat(res.ReportPath); err != nil {
		t.Fatalf("report.md missing: %v", err)
	}
}

func TestExecute_NoChangesIsErrorForFeatureMode(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	bin := buildFakeAgent(t)
	adapter := yaml.New(fakeSpec(bin))
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "do nothing useful",
		Mode:     ModeFeature,
		Autonomy: AutonomySafeExecute,
	})
	if err == nil {
		t.Fatal("expected error for no_changes in feature mode")
	}
	if res.Status != StatusNoChanges {
		t.Fatalf("expected no_changes, got %s", res.Status)
	}
}

func TestExecute_AgentFailureRecorded(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	bin := buildFakeAgent(t)
	adapter := yaml.New(fakeSpec(bin))
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "please fail this run intentionally",
		Mode:     ModeFeature,
		Autonomy: AutonomySafeExecute,
	})
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if res.Status != StatusAgentFailed {
		t.Fatalf("expected agent_failed, got %s", res.Status)
	}
	if !strings.Contains(strings.ToLower(res.ErrorMessage+res.ErrorType), "fail") && res.ErrorType == "" {
		t.Logf("status correct, error fields: type=%q msg=%q", res.ErrorType, res.ErrorMessage)
	}
}

func TestExecute_ApplyMergesIntoProjectRoot(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	bin := buildFakeAgent(t)
	adapter := yaml.New(fakeSpec(bin))
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "create hello.txt with content: world",
		Mode:     ModeFeature,
		Apply:    true,
		Autonomy: AutonomySafeExecute,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Status != StatusApplied {
		t.Fatalf("expected applied, got %s (err=%s)", res.Status, res.ErrorMessage)
	}
	if _, err := os.Stat(filepath.Join(root, "hello.txt")); err != nil {
		t.Fatalf("hello.txt missing after apply: %v", err)
	}
}

func TestExecute_HighRiskRequiresApprovalUnderSafeExecute(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	bin := buildFakeAgent(t)
	adapter := yaml.New(fakeSpec(bin))
	ex := NewDefaultExecutor(root, adapter, nil, index.Profile{})
	res, err := ex.Execute(context.Background(), Request{
		Prompt:   "create Dockerfile with content: FROM scratch",
		Mode:     ModeFeature,
		Apply:    true,
		Autonomy: AutonomySafeExecute,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Status != StatusWaitingApproval {
		t.Fatalf("expected waiting_approval for high risk under safe_execute, got %s", res.Status)
	}
}
