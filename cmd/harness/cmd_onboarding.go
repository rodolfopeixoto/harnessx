// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/ui"
	"github.com/ropeixoto/harnessx/internal/version"
)

func newOnboardingCmd() *cobra.Command {
	var interactive bool
	c := &cobra.Command{
		Use:   "onboarding",
		Short: "Detect installed agent CLIs + dev tools and print next-step recipe",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			if interactive {
				return runOnboardingInteractive(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), dir)
			}
			return runOnboarding(cmd.Context(), cmd.OutOrStdout(), dir)
		},
	}
	c.Flags().BoolVar(&interactive, "interactive", false, "prompt to pin a suggested adapter and run setup actions")
	return c
}

func runOnboardingInteractive(ctx context.Context, in io.Reader, out io.Writer, root string) error {
	r := gatherOnboarding(ctx, root)
	renderOnboarding(out, r)
	reader := bufio.NewReader(in)

	offerMissingToolInstalls(ctx, reader, out, r.Tools)
	if pinErr := offerAdapterPin(ctx, reader, out, root, r.Adapters); pinErr != nil {
		return pinErr
	}
	offerScaffold(ctx, reader, out)
	fmt.Fprintln(out, "\n"+ui.Heading.Render("setup done")+" — open a chat with "+ui.Accent.Render("harness chat --auto-gate"))
	return nil
}

func offerMissingToolInstalls(ctx context.Context, in *bufio.Reader, out io.Writer, tools []checkedTool) {
	missing := []checkedTool{}
	for _, t := range tools {
		if !t.found && strings.HasPrefix(t.installHint, "brew install ") {
			missing = append(missing, t)
		}
	}
	if len(missing) == 0 {
		return
	}
	fmt.Fprintln(out, "\n"+ui.Heading.Render("missing system tools"))
	for _, t := range missing {
		fmt.Fprintf(out, "  %s missing — install via %s ?\n", ui.Accent.Render(t.name), ui.Muted.Render(t.installHint))
		if !askYesNo(in, out, "  install now?", true) {
			fmt.Fprintln(out, "  skipped "+t.name)
			continue
		}
		args := strings.Fields(strings.TrimPrefix(t.installHint, "brew install "))
		bargs := append([]string{"install"}, args...)
		fmt.Fprintln(out, "  → brew "+strings.Join(bargs, " "))
		c := exec.CommandContext(ctx, "brew", bargs...)
		c.Stdout = out
		c.Stderr = out
		if err := c.Run(); err != nil {
			fmt.Fprintf(out, "  %s brew install %s failed: %v\n", ui.MarkFail(), t.name, err)
			continue
		}
		fmt.Fprintln(out, "  "+ui.MarkSuccess()+" installed "+t.name)
	}
}

func offerAdapterPin(ctx context.Context, in *bufio.Reader, out io.Writer, root string, adapters []checkedTool) error {
	installed := []string{}
	for _, a := range adapters {
		if a.found {
			installed = append(installed, a.name)
		}
	}
	fmt.Fprintln(out, "\n"+ui.Heading.Render("adapter pin"))
	if len(installed) == 0 {
		fmt.Fprintln(out, "  "+ui.MarkWarn()+" no agent CLI on PATH; install one of: claude / codex / gemini / kimi / ollama")
		return nil
	}
	choice, err := askChoice(in, out, "  pick the default adapter:", append(installed, "skip"), 0)
	if err != nil {
		return err
	}
	if choice == len(installed) {
		fmt.Fprintln(out, "  skipped pin — use `harness use <id>` later")
		return nil
	}
	pick := installed[choice]
	if err := runHarnessSubcommand(ctx, out, root, "use", pick); err != nil {
		return fmt.Errorf("onboarding: pin %s: %w", pick, err)
	}
	fmt.Fprintln(out, "  "+ui.MarkSuccess()+" pinned "+ui.Accent.Render(pick))
	return nil
}

func offerScaffold(ctx context.Context, in *bufio.Reader, out io.Writer) {
	fmt.Fprintln(out, "\n"+ui.Heading.Render("scaffold a starter project?"))
	stacks := []string{
		"python-ecommerce (FastAPI todo skeleton)",
		"rails (Rails 7 API skeleton)",
		"go (net/http server)",
		"rust (Axum web server)",
		"ruby (Sinatra app)",
		"react (Vite + Vitest)",
		"skip",
	}
	idx, err := askChoice(in, out, "  pick a stack:", stacks, len(stacks)-1)
	if err != nil {
		return
	}
	if idx == len(stacks)-1 {
		fmt.Fprintln(out, "  skipped scaffold — run `harness new <stack> ./dir --yes --with-deps` later")
		return
	}
	stackID := strings.Fields(stacks[idx])[0]
	target, ok := askLine(in, out, "  target directory", "./my-"+stackID)
	if !ok {
		return
	}
	withDeps := askYesNo(in, out, "  install dependencies (--with-deps)?", true)
	args := []string{"new", stackID, target, "--yes"}
	if withDeps {
		args = append(args, "--with-deps")
	}
	if err := runHarnessSubcommand(ctx, out, ".", args...); err != nil {
		fmt.Fprintf(out, "  %s scaffold failed: %v\n", ui.MarkFail(), err)
		return
	}
	fmt.Fprintf(out, "  %s scaffold ready at %s\n", ui.MarkSuccess(), ui.Accent.Render(target))
	fmt.Fprintf(out, "  next: %s\n", ui.Muted.Render("cd "+target+" && harness chat --auto-gate"))
}

func askYesNo(in io.Reader, out io.Writer, prompt string, defaultYes bool) bool {
	def := "Y/n"
	if !defaultYes {
		def = "y/N"
	}
	fmt.Fprintf(out, "  %s [%s]: ", prompt, def)
	answer := trimAndLower(readPromptLine(in))
	if answer == "" {
		return defaultYes
	}
	switch answer {
	case "y", "yes", "1", "true", "ok":
		return true
	case "n", "no", "0", "false":
		return false
	}
	return defaultYes
}

func askChoice(in io.Reader, out io.Writer, prompt string, options []string, defaultIdx int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("askChoice: no options")
	}
	fmt.Fprintln(out, prompt)
	for i, o := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "→ "
		}
		fmt.Fprintf(out, "  %s%d) %s\n", marker, i+1, o)
	}
	fmt.Fprintf(out, "  pick [1-%d, default %d]: ", len(options), defaultIdx+1)
	answer := trimAndLower(readPromptLine(in))
	if answer == "" {
		return defaultIdx, nil
	}
	n, err := strconv.Atoi(answer)
	if err != nil || n < 1 || n > len(options) {
		fmt.Fprintf(out, "  %s invalid choice %q — using default %d\n", ui.MarkWarn(), answer, defaultIdx+1)
		return defaultIdx, nil
	}
	return n - 1, nil
}

func askLine(in io.Reader, out io.Writer, prompt, defaultVal string) (string, bool) {
	fmt.Fprintf(out, "  %s [%s]: ", prompt, defaultVal)
	answer := strings.TrimSpace(readPromptLine(in))
	if answer == "" {
		return defaultVal, true
	}
	return answer, true
}

func readPromptLine(in io.Reader) string {
	if br, ok := in.(*bufio.Reader); ok {
		line, _ := br.ReadString('\n')
		return strings.TrimRight(line, "\r\n")
	}
	buf := make([]byte, 256)
	n, _ := in.Read(buf)
	answer := ""
	for i := 0; i < n; i++ {
		if buf[i] == '\n' || buf[i] == '\r' {
			break
		}
		answer += string(buf[i])
	}
	return answer
}

func trimAndLower(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out = append(out, c)
	}
	return string(out)
}

func runHarnessSubcommand(ctx context.Context, out io.Writer, dir string, args ...string) error {
	bin, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
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
