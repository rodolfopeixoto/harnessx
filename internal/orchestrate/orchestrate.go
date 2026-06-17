// SPDX-License-Identifier: MIT

// Package orchestrate implements paper §4.1.1 functional role
// specialization + §4.1.3 chain/cyclic topologies + §4.3.1 file-only
// blackboard substrate. A flow declares roles (Manager/Planner/Coder/
// Reviewer/Tester) connected by topology; each role contributes a
// structured entry to a shared blackboard.json read by the next role.
package orchestrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
)

type Role string

const (
	RoleManager  Role = "manager"
	RolePlanner  Role = "planner"
	RoleCoder    Role = "coder"
	RoleReviewer Role = "reviewer"
	RoleTester   Role = "tester"
)

func KnownRoles() []Role {
	return []Role{RoleManager, RolePlanner, RoleCoder, RoleReviewer, RoleTester}
}

func validRole(r Role) bool {
	for _, k := range KnownRoles() {
		if r == k {
			return true
		}
	}
	return false
}

type Topology string

const (
	TopologyChain  Topology = "chain"
	TopologyCyclic Topology = "cyclic"
)

func validTopology(t Topology) bool { return t == TopologyChain || t == TopologyCyclic }

type Step struct {
	Role    Role     `yaml:"role"`
	Adapter string   `yaml:"adapter"`
	Command []string `yaml:"command"`
	Prompt  string   `yaml:"prompt"`
}

type Flow struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Topology    Topology `yaml:"topology"`
	MaxCycles   int      `yaml:"max_cycles"`
	Steps       []Step   `yaml:"steps"`
}

func (f Flow) Validate() error {
	if f.Name == "" {
		return errors.New("orchestrate: flow missing name")
	}
	if !validTopology(f.Topology) {
		return fmt.Errorf("orchestrate: invalid topology %q", f.Topology)
	}
	if f.Topology == TopologyCyclic && f.MaxCycles <= 0 {
		return errors.New("orchestrate: cyclic flow requires max_cycles >= 1")
	}
	if len(f.Steps) == 0 {
		return errors.New("orchestrate: flow has no steps")
	}
	for i, s := range f.Steps {
		if !validRole(s.Role) {
			return fmt.Errorf("orchestrate: step %d: unknown role %q (want %v)", i, s.Role, KnownRoles())
		}
		if len(s.Command) == 0 && s.Adapter == "" {
			return fmt.Errorf("orchestrate: step %d (%s): needs command or adapter", i, s.Role)
		}
	}
	return nil
}

func LoadFile(path string) (Flow, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Flow{}, err
	}
	var f Flow
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return Flow{}, fmt.Errorf("orchestrate: parse %s: %w", path, err)
	}
	if err := f.Validate(); err != nil {
		return Flow{}, err
	}
	return f, nil
}

func List(root string) ([]string, error) {
	d := filepath.Join(root, ".harness", "orchestrations")
	entries, err := os.ReadDir(d)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext := filepath.Ext(e.Name()); ext == ".yaml" || ext == ".yml" {
			names = append(names, e.Name()[:len(e.Name())-len(ext)])
		}
	}
	sort.Strings(names)
	return names, nil
}

func Load(root, name string) (Flow, error) {
	for _, ext := range []string{".yaml", ".yml"} {
		p := filepath.Join(root, ".harness", "orchestrations", name+ext)
		if _, err := os.Stat(p); err == nil {
			return LoadFile(p)
		}
	}
	return Flow{}, fmt.Errorf("orchestrate: flow %q not found under .harness/orchestrations/", name)
}

type BlackboardEntry struct {
	Step    int       `json:"step"`
	Role    Role      `json:"role"`
	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
	Stdout  string    `json:"stdout,omitempty"`
	Stderr  string    `json:"stderr,omitempty"`
	Status  string    `json:"status"`
}

type RunResult struct {
	RunID    string            `json:"run_id"`
	Flow     string            `json:"flow"`
	Topology Topology          `json:"topology"`
	Entries  []BlackboardEntry `json:"entries"`
	OK       bool              `json:"ok"`
}

type RunOptions struct {
	Root          string
	Flow          Flow
	DryRun        bool
	StepTimeout   time.Duration
	AdapterRunner AdapterRunner
}

type AdapterRunner func(ctx context.Context, step Step, prevBlackboard []BlackboardEntry) (string, error)

func Run(ctx context.Context, opts RunOptions, out io.Writer) (RunResult, error) {
	if err := opts.Flow.Validate(); err != nil {
		return RunResult{}, err
	}
	runID := ids.New()
	runDir := filepath.Join(opts.Root, ".harness", "artifacts", "runs", runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return RunResult{}, err
	}
	res := RunResult{RunID: runID, Flow: opts.Flow.Name, Topology: opts.Flow.Topology, OK: true}
	cycles := 1
	if opts.Flow.Topology == TopologyCyclic {
		cycles = opts.Flow.MaxCycles
	}
	timeout := opts.StepTimeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	stepIdx := 0
	for c := 0; c < cycles; c++ {
		for i, step := range opts.Flow.Steps {
			entry := BlackboardEntry{Step: stepIdx, Role: step.Role, Started: time.Now().UTC()}
			fmt.Fprintf(out, "orchestrate: step %d cycle %d/%d role=%s\n", stepIdx, c+1, cycles, step.Role)
			if opts.DryRun {
				entry.Status = "dry-run"
			} else if len(step.Command) > 0 {
				stdout, stderr, err := runStep(ctx, opts.Root, step.Command, timeout)
				entry.Stdout = truncate(stdout, 8_000)
				entry.Stderr = truncate(stderr, 4_000)
				if err != nil {
					entry.Status = "fail"
					res.OK = false
				} else {
					entry.Status = "ok"
				}
			} else if step.Adapter != "" && opts.AdapterRunner != nil {
				stdout, err := opts.AdapterRunner(ctx, step, res.Entries)
				entry.Stdout = truncate(stdout, 8_000)
				if err != nil {
					entry.Status = "fail"
					entry.Stderr = err.Error()
					res.OK = false
				} else {
					entry.Status = "ok"
				}
			} else {
				entry.Status = "adapter-step-not-executed-yet"
			}
			entry.Ended = time.Now().UTC()
			res.Entries = append(res.Entries, entry)
			stepIdx++
			if !res.OK && opts.Flow.Topology == TopologyChain {
				break
			}
			_ = i
		}
	}
	if err := writeBlackboard(runDir, res); err != nil {
		return res, err
	}
	fmt.Fprintf(out, "orchestrate: blackboard %s\n", filepath.Join(runDir, "blackboard.json"))
	return res, nil
}

func runStep(ctx context.Context, root string, cmd []string, timeout time.Duration) (string, string, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c := exec.CommandContext(cctx, cmd[0], cmd[1:]...)
	c.Dir = root
	c.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	stdout, err := c.Output()
	stderr := ""
	if ee, ok := err.(*exec.ExitError); ok {
		stderr = string(ee.Stderr)
	}
	return string(stdout), stderr, err
}

func writeBlackboard(dir string, res RunResult) error {
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "blackboard.json"), out, 0o644)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
