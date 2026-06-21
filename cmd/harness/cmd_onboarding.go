// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/ui"
	"github.com/ropeixoto/harnessx/internal/version"
)

func newOnboardingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "onboarding",
		Short: "Detect installed agent CLIs + dev tools and print next-step recipe",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			return runOnboarding(cmd.Context(), cmd.OutOrStdout(), dir)
		},
	}
}

type toolCheck struct {
	name        string
	binary      string
	purpose     string
	installHint string
}

var onboardingTools = []toolCheck{
	{"git", "git", "version control (required)", "brew install git"},
	{"python3", "python3", "python scaffolds + venv", "brew install python"},
	{"node", "node", "react/typescript scaffolds", "brew install node"},
	{"uv", "uv", "python dep installer (used by --with-deps)", "brew install uv"},
	{"rg", "rg", "secrets_scan + context provider", "brew install ripgrep"},
	{"jq", "jq", "tutorial curl examples", "brew install jq"},
	{"rclone", "rclone", "harness backup snapshot", "brew install rclone"},
}

var onboardingAdapters = []toolCheck{
	{"claude", "claude", "Anthropic Claude Code CLI", "https://docs.claude.com/en/docs/claude-code"},
	{"codex", "codex", "OpenAI Codex CLI", "https://github.com/openai/codex"},
	{"gemini", "gemini", "Google Gemini CLI", "https://github.com/google-gemini/gemini-cli"},
	{"kimi", "kimi", "Moonshot Kimi CLI", "https://platform.moonshot.cn/"},
	{"ollama", "ollama", "Local Ollama runtime", "https://ollama.com/"},
}

type onboardingResult struct {
	Tools     []checkedTool
	Adapters  []checkedTool
	HarnessV  string
	Suggested string
}

type checkedTool struct {
	toolCheck
	found   bool
	version string
}

func runOnboarding(ctx context.Context, out io.Writer, root string) error {
	result := gatherOnboarding(ctx, root)
	renderOnboarding(out, result)
	return nil
}

func gatherOnboarding(ctx context.Context, root string) onboardingResult {
	r := onboardingResult{HarnessV: version.Version}
	for _, t := range onboardingTools {
		r.Tools = append(r.Tools, probeTool(ctx, t))
	}
	for _, t := range onboardingAdapters {
		r.Adapters = append(r.Adapters, probeTool(ctx, t))
	}
	r.Suggested = pickSuggestedAdapter(r.Adapters, root)
	return r
}

func probeTool(ctx context.Context, t toolCheck) checkedTool {
	c := checkedTool{toolCheck: t}
	path, err := exec.LookPath(t.binary)
	if err != nil || path == "" {
		return c
	}
	c.found = true
	if v := probeVersion(ctx, t.binary); v != "" {
		c.version = v
	}
	return c
}

func probeVersion(ctx context.Context, binary string) string {
	cmd := exec.CommandContext(ctx, binary, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	line := string(out)
	if idx := indexNL(line); idx > 0 {
		line = line[:idx]
	}
	return trimToWidth(line, 60)
}

func indexNL(s string) int {
	for i, r := range s {
		if r == '\n' || r == '\r' {
			return i
		}
	}
	return -1
}

func trimToWidth(s string, w int) string {
	if len(s) <= w {
		return s
	}
	return s[:w] + "…"
}

func pickSuggestedAdapter(adapters []checkedTool, root string) string {
	reg, _, err := agentcmd.LoadAll(root)
	if err == nil {
		ids := reg.IDs()
		for _, preferred := range []string{"claude", "codex", "kimi", "gemini", "ollama"} {
			for _, id := range ids {
				if id == preferred {
					for _, a := range adapters {
						if a.name == preferred && a.found {
							return preferred
						}
					}
				}
			}
		}
	}
	for _, a := range adapters {
		if a.found {
			return a.name
		}
	}
	return ""
}

func renderOnboarding(out io.Writer, r onboardingResult) {
	fmt.Fprintln(out, ui.Heading.Render("harness onboarding"))
	fmt.Fprintf(out, "  %s harness   %s\n", ui.Muted.Render("·"), ui.Accent.Render(r.HarnessV))

	fmt.Fprintln(out, "\n"+ui.Heading.Render("system tools"))
	for _, t := range r.Tools {
		renderTool(out, t)
	}

	fmt.Fprintln(out, "\n"+ui.Heading.Render("agent adapters"))
	for _, t := range r.Adapters {
		renderTool(out, t)
	}

	fmt.Fprintln(out, "\n"+ui.Heading.Render("next steps"))
	if r.Suggested != "" {
		fmt.Fprintf(out, "  1. pin an agent:   %s\n", ui.Accent.Render("harness use "+r.Suggested))
	} else {
		fmt.Fprintf(out, "  1. %s no agent CLI on PATH — install one above to enable chat\n", ui.MarkWarn())
	}
	fmt.Fprintf(out, "  2. scaffold:        %s\n", ui.Accent.Render("harness new python-ecommerce ./my-api --yes --with-deps"))
	fmt.Fprintf(out, "  3. open chat:       %s\n", ui.Accent.Render("cd my-api && harness chat --auto-gate"))
	fmt.Fprintf(out, "  4. drive a feature: %s\n", ui.Muted.Render(`/drive add /healthz with pytest`))
	fmt.Fprintf(out, "  5. tutorial:        %s\n", ui.Muted.Render("docs/TUTORIAL-TODOIST.md"))
}

func renderTool(out io.Writer, c checkedTool) {
	mark := ui.MarkFail()
	if c.found {
		mark = ui.MarkSuccess()
	}
	line := fmt.Sprintf("  %s %-10s", mark, c.name)
	if c.found {
		line += "  " + ui.Muted.Render(c.version)
	} else {
		line += "  " + ui.Muted.Render(c.purpose+" — install: "+c.installHint)
	}
	fmt.Fprintln(out, line)
}
