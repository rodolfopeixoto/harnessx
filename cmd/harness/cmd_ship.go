// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/plancontract"
	"github.com/ropeixoto/harnessx/internal/scm"
	"github.com/ropeixoto/harnessx/internal/sensors/planscope"
)

type shipOptions struct {
	prompt         string
	maxAttempts    int
	rateLimitWait  time.Duration
	rateLimitTries int
	branchBase     string
	branchPrefix   string
	autonomy       string
	dryRun         bool
	skipCommit     bool
	planID         string
	agentID        string
	watch          bool
	watchInterval  time.Duration
	allowDirty     bool
	budgetUSD      float64
}

func newShipCmd() *cobra.Command {
	opts := shipOptions{
		maxAttempts:    3,
		rateLimitWait:  20 * time.Second,
		rateLimitTries: 5,
		branchBase:     "develop",
		branchPrefix:   "feature",
		autonomy:       "ask",
		budgetUSD:      1.0,
	}
	c := &cobra.Command{
		Use:   "ship <prompt>",
		Short: "Single-command SDLC driver: branch → spec → do → ci-loop → commit",
		Long: `Drives the full development cycle from one prompt:

  1. ensure clean git tree on the base branch (default 'develop')
  2. branch feature/<slug>
  3. write spec via 'harness feature'
  4. up to --max-attempts iterations of 'harness do' + 'harness ci'
  5. conventional commit on success

The loop retries on transient adapter errors (HTTP 429, rate-limit
markers in stderr) using --rate-limit-retries x --rate-limit-wait
backoff before falling through to the next adapter in the router
chain.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.prompt = strings.Join(args, " ")
			if opts.watch {
				return runShipWatch(cmd.Context(), cmd.OutOrStdout(), opts)
			}
			return runShip(cmd.Context(), cmd.OutOrStdout(), opts)
		},
	}
	c.Flags().IntVar(&opts.maxAttempts, "max-attempts", opts.maxAttempts, "max do→ci iterations")
	c.Flags().DurationVar(&opts.rateLimitWait, "rate-limit-wait", opts.rateLimitWait, "initial backoff for 429 retries")
	c.Flags().IntVar(&opts.rateLimitTries, "rate-limit-retries", opts.rateLimitTries, "retries before failing a step on 429")
	c.Flags().StringVar(&opts.branchBase, "base", opts.branchBase, "base branch to branch from")
	c.Flags().StringVar(&opts.branchPrefix, "branch-prefix", opts.branchPrefix, "branch prefix (feature|fix|chore)")
	c.Flags().StringVar(&opts.autonomy, "autonomy", opts.autonomy, "autonomy mode forwarded to 'harness do'")
	c.Flags().BoolVar(&opts.dryRun, "dry-run", false, "print steps without invoking them")
	c.Flags().BoolVar(&opts.skipCommit, "skip-commit", false, "do not create the final conventional commit")
	c.Flags().StringVar(&opts.planID, "plan", "", "PLAN contract id (ulid or filename); paper §3.4.2")
	c.Flags().StringVar(&opts.agentID, "agent", "", "force a specific adapter id (overrides router + active pin)")
	c.Flags().BoolVar(&opts.watch, "watch", false, "re-run ship loop whenever a project file changes")
	c.Flags().DurationVar(&opts.watchInterval, "watch-interval", 3*time.Second, "polling interval in --watch mode")
	c.Flags().BoolVar(&opts.allowDirty, "allow-dirty", false, "do not require a clean working tree before shipping")
	c.Flags().Float64Var(&opts.budgetUSD, "budget-usd", opts.budgetUSD, "max USD spent across the ship loop; forwarded to 'harness do'")
	// --yes is accepted as a no-op so `harness ship` can be invoked the same
	// way every other harness command is (`harness do --yes`, `harness new
	// --yes`). The REPL /ship slash appends it automatically.
	var yesNoop bool
	c.Flags().BoolVar(&yesNoop, "yes", false, "accepted for parity with other harness commands; no prompts to skip")
	_ = yesNoop
	return c
}

func runShip(ctx context.Context, out io.Writer, opts shipOptions) error {
	root, err := prepareShip(ctx, out, &opts)
	if err != nil {
		return err
	}
	slug := slugify(opts.prompt)
	branch := fmt.Sprintf("%s/%s", opts.branchPrefix, slug)
	bin, err := os.Executable()
	if err != nil {
		return err
	}
	steps := buildShipSteps(ctx, out, bin, root, opts, branch)
	return runShipSteps(ctx, out, root, opts, branch, steps)
}

func prepareShip(ctx context.Context, out io.Writer, opts *shipOptions) (string, error) {
	root, err := cwd()
	if err != nil {
		return "", err
	}
	if !scm.HasRepo(root) {
		return "", errors.New("ship: not a git repo (run 'git init' or 'harness init --git' first)")
	}
	dirty, err := gitDirty(ctx, root)
	if err != nil {
		return "", err
	}
	if dirty && !opts.dryRun && !opts.allowDirty {
		return "", errors.New("ship: working tree dirty; commit, stash, or pass --allow-dirty")
	}
	if dirty && opts.allowDirty {
		fmt.Fprintln(out, "ship: working tree dirty — proceeding because --allow-dirty was set")
	}
	if opts.planID != "" {
		contract, err := plancontract.Load(root, opts.planID)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(out, "ship: plan %s loaded — files=%d invariants=%d risk=%s\n",
			contract.ID, len(contract.Files), len(contract.Invariants), contract.Risk)
		if opts.prompt == "" || strings.HasPrefix(strings.ToLower(opts.prompt), "plan:") {
			opts.prompt = contract.Intent
		}
	}
	opts.agentID = activeagent.ResolveAgentID(root, opts.agentID)
	if opts.agentID != "" {
		fmt.Fprintf(out, "ship: agent=%s\n", opts.agentID)
	}
	return root, nil
}

func buildShipSteps(ctx context.Context, out io.Writer, bin, root string, opts shipOptions, branch string) []func() error {
	steps := []func() error{
		func() error { return shipCheckoutBranch(ctx, out, root, opts, branch) },
		func() error {
			return shipRunStep(ctx, out, bin, root, opts, "feature spec",
				[]string{"feature", opts.prompt, "--yes"})
		},
	}
	for i := 0; i < opts.maxAttempts; i++ {
		attempt := i + 1
		steps = append(steps,
			func() error {
				doArgs := []string{"do", opts.prompt, "--autonomy", opts.autonomy, "--yes"}
				if opts.agentID != "" {
					doArgs = append(doArgs, "--agent", opts.agentID)
				}
				if opts.budgetUSD > 0 {
					doArgs = append(doArgs, "--budget-usd", strconv.FormatFloat(opts.budgetUSD, 'f', -1, 64))
				}
				return shipRunStep(ctx, out, bin, root, opts,
					fmt.Sprintf("do attempt %d", attempt), doArgs)
			},
			func() error {
				err := shipRunStep(ctx, out, bin, root, opts, fmt.Sprintf("ci attempt %d", attempt),
					[]string{"ci"})
				if err == nil {
					return errShipCISucceeded
				}
				if attempt == opts.maxAttempts {
					return fmt.Errorf("ship: ci still red after %d attempts: %w", attempt, err)
				}
				fmt.Fprintf(out, "ship: ci red on attempt %d; retrying\n", attempt)
				return nil
			},
		)
	}
	if !opts.skipCommit {
		steps = append(steps, func() error {
			return shipCommit(ctx, out, root, opts, branch)
		})
	}
	return steps
}

func runShipSteps(ctx context.Context, out io.Writer, root string, opts shipOptions, branch string, steps []func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			if errors.Is(err, errShipCISucceeded) {
				if opts.skipCommit {
					fmt.Fprintln(out, "ship: ci green — skipping commit per --skip-commit")
					return nil
				}
				return shipCommit(ctx, out, root, opts, branch)
			}
			return err
		}
	}
	return nil
}

var errShipCISucceeded = errors.New("ci succeeded")

func shipCheckoutBranch(ctx context.Context, out io.Writer, root string, opts shipOptions, branch string) error {
	fmt.Fprintf(out, "ship: branch %s ← %s\n", branch, opts.branchBase)
	if opts.dryRun {
		return nil
	}
	if err := runGit(ctx, root, "fetch", "--quiet", "origin"); err != nil {
		fmt.Fprintf(out, "ship: fetch skipped (%v)\n", err)
	}
	if err := runGit(ctx, root, "checkout", opts.branchBase); err != nil {
		fmt.Fprintf(out, "ship: base %s missing; staying on current branch\n", opts.branchBase)
	}
	return runGit(ctx, root, "checkout", "-B", branch)
}

func shipRunStep(ctx context.Context, out io.Writer, bin, root string, opts shipOptions, label string, args []string) error {
	fmt.Fprintf(out, "ship: %s — harness %s\n", label, strings.Join(args, " "))
	if opts.dryRun {
		return nil
	}
	wait := opts.rateLimitWait
	for attempt := 0; attempt <= opts.rateLimitTries; attempt++ {
		err := runHarness(ctx, bin, root, out, args)
		if err == nil {
			return nil
		}
		if !isRateLimit(err.Error()) {
			return err
		}
		if attempt == opts.rateLimitTries {
			return fmt.Errorf("rate-limit budget exhausted: %w", err)
		}
		fmt.Fprintf(out, "ship: rate-limit hit; sleeping %s before retry %d/%d\n", wait, attempt+1, opts.rateLimitTries)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		wait *= 2
	}
	return nil
}

func shipCommit(ctx context.Context, out io.Writer, root string, opts shipOptions, branch string) error {
	fmt.Fprintf(out, "ship: commit on %s\n", branch)
	if opts.dryRun {
		return nil
	}
	if opts.planID != "" {
		check, err := planscope.Check(ctx, planscope.Options{Root: root, PlanID: opts.planID})
		if err != nil {
			return fmt.Errorf("ship: scope check: %w", err)
		}
		fmt.Fprint(out, planscope.FormatResult(check))
		if !check.Pass() {
			return fmt.Errorf("ship: scope violations against PLAN-%s; refusing commit", check.PlanID)
		}
	}
	if err := runGit(ctx, root, "add", "-A"); err != nil {
		return err
	}
	subject := conventionalSubject(opts.branchPrefix, opts.prompt)
	body := fmt.Sprintf("Generated by `harness ship`.\n\nPrompt: %s", opts.prompt)
	if opts.planID != "" {
		body += "\nPlan: " + opts.planID
	}
	return runGit(ctx, root, "commit", "-m", subject, "-m", body)
}

func runHarness(ctx context.Context, bin, root string, out io.Writer, args []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "HARNESS_PLAIN=1", "NO_COLOR=1")
	var stderr bytes.Buffer
	cmd.Stdout = out
	cmd.Stderr = io.MultiWriter(out, &stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}
	return nil
}

func runGit(ctx context.Context, root string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, string(out))
	}
	return nil
}

func gitDirty(ctx context.Context, root string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

var rateLimitRe = regexp.MustCompile(`(?i)(429|rate[\s_-]?limit|too many requests|quota.*exceed)`)

func isRateLimit(s string) bool { return rateLimitRe.MatchString(s) }

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
		s = strings.Trim(s, "-")
	}
	if s == "" {
		s = "change"
	}
	return s
}

func conventionalSubject(prefix, prompt string) string {
	conv := "feat"
	switch prefix {
	case "fix", "hotfix":
		conv = "fix"
	case "chore":
		conv = "chore"
	case "refactor":
		conv = "refactor"
	case "docs":
		conv = "docs"
	}
	short := strings.TrimSpace(prompt)
	if len(short) > 50-len(conv)-2 {
		short = short[:50-len(conv)-2]
	}
	return fmt.Sprintf("%s: %s", conv, short)
}
