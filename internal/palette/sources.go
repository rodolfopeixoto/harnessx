// SPDX-License-Identifier: MIT

package palette

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/workspace"
)

const (
	SourceProjects     = "projects"
	SourceCapabilities = "capabilities"
	SourceCommands     = "commands"
)

type ProjectsSource struct {
	Registry *workspace.Registry
}

func (p ProjectsSource) Name() string { return SourceProjects }

func (p ProjectsSource) Search(ctx context.Context, q string) ([]Hit, error) {
	if p.Registry == nil {
		return nil, nil
	}
	projects, err := p.Registry.List(ctx, true)
	if err != nil {
		return nil, err
	}
	var out []Hit
	for _, proj := range projects {
		score := bestScore(q, proj.Slug, proj.DisplayName, proj.RootPath)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceProjects,
			Kind:       "project",
			Title:      proj.DisplayName,
			Subtitle:   proj.RootPath,
			RouterPath: "/workspace/" + proj.Slug,
			Score:      score,
		})
	}
	return out, nil
}

type CapabilitiesSource struct {
	Catalog *catalog.Catalog
	Root    string
}

func (c CapabilitiesSource) Name() string { return SourceCapabilities }

func (c CapabilitiesSource) Search(ctx context.Context, q string) ([]Hit, error) {
	if c.Catalog == nil {
		return nil, nil
	}
	caps, err := c.Catalog.Discover(ctx, c.Root)
	if err != nil {
		return nil, err
	}
	var out []Hit
	for _, cap := range caps {
		score := bestScore(q, cap.Name, string(cap.Kind), cap.Description)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceCapabilities,
			Kind:       string(cap.Kind),
			Title:      string(cap.Kind) + "/" + cap.Name,
			Subtitle:   cap.Description,
			RouterPath: "/catalog/" + string(cap.Kind) + "/" + cap.Name,
			Score:      score,
		})
	}
	_ = domain.AllCapabilityKinds()
	return out, nil
}

type CommandsSource struct {
	Commands []Command
}

type Command struct {
	Name        string
	Description string
	RouterPath  string
}

func (c CommandsSource) Name() string { return SourceCommands }

func (c CommandsSource) Search(_ context.Context, q string) ([]Hit, error) {
	var out []Hit
	for _, cmd := range c.Commands {
		score := bestScore(q, cmd.Name, cmd.Description)
		if score == 0 {
			continue
		}
		out = append(out, Hit{
			Source:     SourceCommands,
			Kind:       "command",
			Title:      cmd.Name,
			Subtitle:   cmd.Description,
			RouterPath: cmd.RouterPath,
			Score:      score,
		})
	}
	return out, nil
}

var BuiltinCommands = []Command{
	{Name: "Open workspace", Description: "Show all registered projects", RouterPath: "/workspace"},
	{Name: "Open catalog", Description: "Browse capabilities (agents, MCPs, hooks, sensors, …)", RouterPath: "/catalog"},
	{Name: "Open sessions", Description: "List recent runs", RouterPath: "/"},
	{Name: "Open settings", Description: "Project + global settings", RouterPath: "/settings"},
	{Name: "version", Description: "Print harness version + build info", RouterPath: "harness:version"},
	{Name: "doctor", Description: "Probe required tools and adapters", RouterPath: "harness:doctor"},
	{Name: "onboarding", Description: "Doctor + next-step recipe", RouterPath: "harness:onboarding"},
	{Name: "ci", Description: "Run the full sensor pack (lint+typecheck+test+security)", RouterPath: "harness:ci"},
	{Name: "check", Description: "Same as ci with a smaller default scope", RouterPath: "harness:check"},
	{Name: "lint", Description: "Run the project's lint command", RouterPath: "harness:lint"},
	{Name: "test", Description: "Run the project's test command", RouterPath: "harness:test"},
	{Name: "coverage", Description: "Stack-aware coverage gate", RouterPath: "harness:coverage"},
	{Name: "config show", Description: "Print effective routing per task type", RouterPath: "harness:config-show"},
	{Name: "config set", Description: "Override default routing for a task", RouterPath: "harness:config-set"},
	{Name: "config wizard", Description: "Interactive routing wizard", RouterPath: "harness:config-wizard"},
	{Name: "route show", Description: "Classify a prompt and print routed adapter chain", RouterPath: "harness:route-show"},
	{Name: "explain", Description: "Same as route show plus reason trace", RouterPath: "harness:explain"},
	{Name: "ask", Description: "Question Mode: evidence gathering (use --agent to synthesize)", RouterPath: "harness:ask"},
	{Name: "plan", Description: "Write deterministic spec + plan template", RouterPath: "harness:plan"},
	{Name: "feature", Description: "Pin Feature Mode for a prompt", RouterPath: "harness:feature"},
	{Name: "bugfix", Description: "Pin Bugfix Mode for a prompt", RouterPath: "harness:bugfix"},
	{Name: "auto", Description: "Plan + execute with autonomy gate", RouterPath: "harness:auto"},
	{Name: "ship", Description: "Plan, execute, sensors, and prepare PR", RouterPath: "harness:ship"},
	{Name: "do", Description: "Free-form prompt: classify, plan, execute", RouterPath: "harness:do"},
	{Name: "runs list", Description: "List persisted runs (newest first)", RouterPath: "harness:runs-list"},
	{Name: "runs inspect", Description: "Show meta + sensors + paths for a run", RouterPath: "harness:runs-inspect"},
	{Name: "runs report", Description: "Print rendered Markdown report for a run", RouterPath: "harness:runs-report"},
	{Name: "runs sensors", Description: "List sensor outcomes for a run", RouterPath: "harness:runs-sensors"},
	{Name: "runs approve", Description: "Approve a waiting_approval run and apply diff", RouterPath: "harness:runs-approve"},
	{Name: "runs discard", Description: "Discard a run's worktree without applying", RouterPath: "harness:runs-discard"},
	{Name: "runs prune", Description: "Garbage-collect old runs", RouterPath: "harness:runs-prune"},
	{Name: "metrics", Description: "Aggregate telemetry across runs", RouterPath: "harness:metrics"},
	{Name: "analytics", Description: "Cross-project session spend", RouterPath: "harness:analytics"},
	{Name: "audit", Description: "Read .harness/audit/events.jsonl", RouterPath: "harness:audit"},
	{Name: "audit-solid", Description: "Run the SOLID audit on Go packages", RouterPath: "harness:audit-solid"},
	{Name: "dependency-audit", Description: "Inventory deps + removal candidates", RouterPath: "harness:dependency-audit"},
	{Name: "image-audit", Description: "Dockerfile findings", RouterPath: "harness:image-audit"},
	{Name: "log-audit", Description: "Noisy log call sites", RouterPath: "harness:log-audit"},
	{Name: "security-audit", Description: "Run every security-category sensor", RouterPath: "harness:security-audit"},
	{Name: "perf-snapshot", Description: "Capture a baseline snapshot for perf-compare", RouterPath: "harness:perf-snapshot"},
	{Name: "perf-compare", Description: "Compare two perf snapshots", RouterPath: "harness:perf-compare"},
	{Name: "memory list", Description: "List durable project memories", RouterPath: "harness:memory-list"},
	{Name: "memory recall", Description: "Score memories against a query", RouterPath: "harness:memory-recall"},
	{Name: "context build", Description: "Build a context pack from a prompt", RouterPath: "harness:context-build"},
	{Name: "context inspect", Description: "Inspect a cached context pack", RouterPath: "harness:context-inspect"},
	{Name: "artifact ls", Description: "List artifacts under .harness/artifacts", RouterPath: "harness:artifact-ls"},
	{Name: "artifact cat", Description: "Print an artifact's contents", RouterPath: "harness:artifact-cat"},
	{Name: "agent list", Description: "List bundled and installed agents", RouterPath: "harness:agent-list"},
	{Name: "agent install", Description: "Install an agent adapter", RouterPath: "harness:agent-install"},
	{Name: "skill list", Description: "List installed skills", RouterPath: "harness:skill-list"},
	{Name: "mcp list", Description: "List detected MCP servers", RouterPath: "harness:mcp-list"},
	{Name: "hook list", Description: "List installed hooks", RouterPath: "harness:hook-list"},
	{Name: "sensor list", Description: "List registered sensors", RouterPath: "harness:sensor-list"},
	{Name: "sensor run", Description: "Run one sensor by id", RouterPath: "harness:sensor-run"},
	{Name: "runtime info", Description: "Show selected container runtime", RouterPath: "harness:runtime-info"},
	{Name: "runtime list", Description: "List candidate container runtimes", RouterPath: "harness:runtime-list"},
	{Name: "runtime set", Description: "Pin the active runtime", RouterPath: "harness:runtime-set"},
	{Name: "containers list", Description: "List running containers", RouterPath: "harness:containers-list"},
	{Name: "containers kill", Description: "Stop a container", RouterPath: "harness:containers-kill"},
	{Name: "containers prune", Description: "Prune stopped containers", RouterPath: "harness:containers-prune"},
	{Name: "images list", Description: "List container images", RouterPath: "harness:images-list"},
	{Name: "images prune", Description: "Prune dangling images", RouterPath: "harness:images-prune"},
	{Name: "secret list", Description: "List configured secrets", RouterPath: "harness:secret-list"},
	{Name: "autonomy get", Description: "Print autonomy gate matrix", RouterPath: "harness:autonomy-get"},
	{Name: "autonomy set", Description: "Set active autonomy level", RouterPath: "harness:autonomy-set"},
	{Name: "health show", Description: "Project health score", RouterPath: "harness:health-show"},
	{Name: "stack tour", Description: "Walk every harness feature against a temp project", RouterPath: "harness:stack-tour"},
	{Name: "stack status", Description: "Probe the dashboard health endpoint", RouterPath: "harness:stack-status"},
	{Name: "stack audit", Description: "Run the visual + functional audit pipeline", RouterPath: "harness:stack-audit"},
	{Name: "dashboard", Description: "Start the local dashboard", RouterPath: "harness:dashboard"},
	{Name: "project current", Description: "Print the active project", RouterPath: "harness:project-current"},
	{Name: "project list", Description: "List registered projects", RouterPath: "harness:project-list"},
	{Name: "scaffold", Description: "Scaffold a new project from a bundled template", RouterPath: "harness:scaffold"},
	{Name: "new", Description: "Interactive project creator", RouterPath: "harness:new"},
	{Name: "init", Description: "Initialize a .harness directory in the cwd", RouterPath: "harness:init"},
	{Name: "install", Description: "Install harness binaries / shells / LSPs", RouterPath: "harness:install"},
	{Name: "update", Description: "Self-update", RouterPath: "harness:update"},
	{Name: "diagnose", Description: "Project diagnostics (stack detection, config sanity)", RouterPath: "harness:diagnose"},
	{Name: "cleanup", Description: "Remove stale harness artifacts", RouterPath: "harness:cleanup"},
	{Name: "use", Description: "Pin the active agent adapter", RouterPath: "harness:use"},
	{Name: "session show", Description: "Print a session's details", RouterPath: "harness:session-show"},
	{Name: "session export", Description: "Export a session as JSON", RouterPath: "harness:session-export"},
	{Name: "completion", Description: "Generate shell completion script", RouterPath: "harness:completion"},
}

func bestScore(q string, candidates ...string) int {
	best := 0
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if s := Score(q, c); s > best {
			best = s
		}
		if s := Score(q, filepath.Base(c)); s > best {
			best = s
		}
	}
	if best == 0 && strings.TrimSpace(q) == "" {
		best = 1
	}
	return best
}
