// SPDX-License-Identifier: MIT

package repl

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/ui"
)

// quickAnswer routes a `/btw <question>` through the cheapest
// adapter chain (task tag cheap_review) so users can fire a side
// question without burning opus tokens. Costs are recorded on the
// turn so /cost + analytics aggregate them under the btw tag.
func quickAnswer(ctx context.Context, sess *Session, opts Options, question string, turn *Turn) {
	if question == "" {
		fmt.Fprintln(opts.Out, "  /btw needs a question")
		return
	}
	if !checkBudget(sess, opts.Out) {
		return
	}
	prompt := "Answer this side question briefly (<=3 sentences). Do not propose code or actions; just a direct answer.\n\nQuestion: " + question
	chatTurnFor(ctx, sess, opts, prompt, "cheap_review", turn)
}

// writerIsTerminal answers true only when out is an *os.File pointing
// at a real TTY. Buffers, pipes, and tee'd log writers all return
// false so the spinner falls back to the static "agent: working…"
// line.
func writerIsTerminal(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// signalAwareCtx returns a derived context that cancels on SIGINT so
// Ctrl-C during a long agent call or `harness ci` aborts just that
// turn instead of killing the whole REPL. Printing the carriage-
// return + clear lets the next prompt land on a fresh line even when
// the spinner was mid-frame. The caller MUST invoke the returned
// cancel func to release the signal handler.
func signalAwareCtx(parent context.Context, out io.Writer) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		select {
		case <-sig:
			fmt.Fprint(out, "\r  \r✗ interrupted — back to prompt (Ctrl-D or /exit to leave)\n")
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(sig)
		cancel()
	}
}

// recapSession asks the cheapest review model to summarise the session
// — typically gemini/kimi according to the cheap_review route — so
// the recap does not burn opus tokens just to bullet-list intent.
func recapSession(ctx context.Context, sess *Session, opts Options, turn *Turn) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(opts.Out, "  no turns to recap yet")
		return
	}
	if opts.Adapter == nil && opts.Route == nil {
		fmt.Fprintln(opts.Out, "  [recap] (no adapter wired — listing inputs)")
		for i, t := range sess.Turns {
			if t.Input == "" {
				continue
			}
			fmt.Fprintf(opts.Out, "  %d. %s\n", i+1, truncateForContext(t.Input, 120))
		}
		return
	}
	prompt := buildRecapPrompt(sess)
	chatTurnFor(ctx, sess, opts, prompt, "cheap_review", turn)
}

func buildRecapPrompt(sess *Session) string {
	var b strings.Builder
	b.WriteString("Summarise this HarnessX chat session in <=8 bullet points. ")
	b.WriteString("Focus on: what was built, what tests/sensors ran, what is still open. ")
	b.WriteString("Do NOT make plans; this is a recap, not a new task.\n\n")
	b.WriteString("# Session turns\n\n")
	for i, t := range sess.Turns {
		in := strings.TrimSpace(t.Input)
		if in == "" {
			continue
		}
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, t.Action, truncateForContext(in, 200))
	}
	return b.String()
}

func executePlan(ctx context.Context, sess *Session, opts Options, prompt string, turn *Turn) {
	if opts.Planner == nil {
		turn.Action = "no-planner"
		fmt.Fprintln(opts.Out, "  ✗ no planner wired — pin an adapter with `harness use <id>` or pass --no-adapter to fall back to deterministic mode")
		return
	}
	plan, err := opts.Planner(ctx, sess.Goal, prompt)
	if err != nil {
		turn.Action = "plan-error"
		fmt.Fprintf(opts.Out, "planner: %v\n", err)
		return
	}
	turn.Plan = &plan
	res, err := intentplan.Execute(ctx, plan, intentplan.ExecOptions{
		HarnessBin: opts.HarnessBin, WorkingDir: opts.Root,
		Out: opts.Out, StepTimeout: opts.StepTimeout,
	})
	if err != nil {
		turn.Action = "execute-error"
		fmt.Fprintf(opts.Out, "executor: %v\n", err)
		return
	}
	turn.Result = &res
	turn.Action = "executed"
	if res.OK {
		fmt.Fprintln(opts.Out, "✓ plan green")
	} else {
		fmt.Fprintln(opts.Out, "✗ plan red — inspect step outputs above")
	}
}

func chatTurn(ctx context.Context, sess *Session, opts Options, prompt string, turn *Turn) {
	chatTurnFor(ctx, sess, opts, prompt, "implementation", turn)
}

// chatTurnFor is the routed variant: when opts.Route is wired and the
// task tag resolves to a registered adapter, we hand the request to
// that adapter instead of opts.Adapter. Lets `/plan` use a cheap
// model while plain text keeps using the implementation chain.
func chatTurnFor(ctx context.Context, sess *Session, opts Options, prompt, task string, turn *Turn) {
	adapter := opts.Adapter
	adapterID := opts.AdapterID
	if opts.Route != nil && opts.RouteEnabled {
		if a, id, err := opts.Route(task); err == nil && a != nil {
			adapter = a
			adapterID = id
		}
	}
	if adapter == nil {
		fmt.Fprintln(opts.Out, "✗ no adapter wired")
		return
	}
	fmt.Fprintf(opts.Out, "  %s %s %s\n",
		ui.Accent.Render("[agent]"),
		ui.Info.Render("calling "+adapterID),
		ui.Muted.Render("("+task+", "+adapterBillingMode(adapterID)+")…"))
	live := &prefixWriter{w: opts.Out, prefix: "  │ "}
	timeout := opts.StepTimeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	spinStop := startSpinner(opts.Out, opts.Plain)
	defer spinStop()
	live.onFirstWrite = spinStop
	res := adapter.Run(rctx, agents.AgentRequest{
		Prompt:     withConversationContext(sess, opts.AdapterID, opts.Model, prompt),
		Model:      opts.Model,
		WorkingDir: opts.Root,
		Timeout:    timeout,
		LiveOut:    live,
		Extra:      map[string]string{"task": task},
	})
	spinStop()
	live.flush()
	if res.Err != nil {
		fmt.Fprintf(opts.Out, "✗ %s: %v\n", adapterID, res.Err)
		return
	}
	if msg := strings.TrimSpace(res.Output.FinalMessage); msg != "" {
		fmt.Fprintln(opts.Out, msg)
	}
	fmt.Fprintf(opts.Out, "  %s %s %s\n",
		ui.MarkSuccess(),
		ui.Accent.Render(adapterID),
		ui.Muted.Render(fmt.Sprintf("done in %s · in=%d out=%d · ~$%.4f",
			res.Duration.Round(time.Millisecond),
			res.Usage.InputTokens, res.Usage.OutputTokens, res.Usage.EstimatedCostUSD)))
	if turn != nil {
		turn.InTokens = res.Usage.InputTokens
		turn.OutTokens = res.Usage.OutputTokens
		turn.CostUSD = res.Usage.EstimatedCostUSD
		turn.AdapterID = adapterID
		turn.TaskTag = task
	}
}

// startSpinner returns a stop function that is idempotent and safe to
// call multiple times. The spinner prints a frame every ~120ms while
// the agent runs so the user can tell the terminal is alive — claude
// and codex CLIs hold their JSON output until the call completes,
// which used to make `harness chat` look frozen for tens of seconds.
func startSpinner(out io.Writer, plain bool) func() {
	if plain || !writerIsTerminal(out) {
		// Pipe / log file / non-TTY: emit a single "agent: working…"
		// line so the consumer still sees liveness without getting
		// CR-clobbered braille glyphs in the log file.
		fmt.Fprintln(out, "agent: working…")
		return func() {}
	}
	frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := time.NewTicker(120 * time.Millisecond)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-stop:
				fmt.Fprint(out, "\r\033[2K")
				return
			case <-t.C:
				fmt.Fprintf(out, "\r\033[2K  %c thinking…", frames[i%len(frames)])
				i++
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			<-done
		})
	}
}

// withConversationContext threads the last few session turns into the
// prompt so multi-turn chat reads as one conversation instead of a
// series of isolated single-shot calls (paper §3.2.1 Working Memory).
// Truncates aggressively because adapter context budgets are bounded
// and the most recent few turns dominate the signal.
func withConversationContext(sess *Session, adapterID, model, prompt string) string {
	identity := buildIdentityPreface(adapterID, model)
	if sess == nil || len(sess.Turns) == 0 {
		return identity + prompt
	}
	var b strings.Builder
	b.WriteString(identity)
	b.WriteString("# Conversation so far\n\n")
	if !writeConversationTurns(&b, sess) {
		return identity + prompt
	}
	b.WriteString("# New user message\n\n")
	b.WriteString(prompt)
	return b.String()
}

// writeConversationTurns appends the trailing window of session turns to
// b and returns true when at least one turn was written. Split out to
// keep withConversationContext's cyclomatic complexity within budget.
func writeConversationTurns(b *strings.Builder, sess *Session) bool {
	const maxTurns = 5
	const maxOutputBytes = 1200
	start := sess.ContextMark
	if len(sess.Turns)-start > maxTurns {
		start = len(sess.Turns) - maxTurns
	}
	if start < 0 {
		start = 0
	}
	wrote := false
	for i := start; i < len(sess.Turns); i++ {
		t := sess.Turns[i]
		in := strings.TrimSpace(t.Input)
		if in == "" || strings.HasPrefix(in, "/") {
			continue
		}
		wrote = true
		fmt.Fprintf(b, "[turn %d] user: %s\n", i+1, in)
		if t.Result != nil && len(t.Result.Steps) > 0 {
			fmt.Fprintf(b, "[turn %d] harness ran %d steps (ok=%t)\n", i+1, len(t.Result.Steps), t.Result.OK)
		}
		if t.Action == "chat" {
			if out := truncateForContext(t.Input, maxOutputBytes); out != "" {
				fmt.Fprintf(b, "[turn %d] agent replied (truncated): %s\n", i+1, out)
			}
		}
		b.WriteByte('\n')
	}
	return wrote
}

// buildIdentityPreface tells the underlying agent which adapter and
// model it is running as. Without it the CLI cannot answer "qual modelo
// está usando?" — most adapters (kimi-cli, antigravity, codex) have no
// introspection of the harness routing layer that picked them.
func buildIdentityPreface(adapterID, model string) string {
	if adapterID == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Runtime identity\n\n")
	fmt.Fprintf(&b, "You are running inside harness chat as the `%s` adapter.\n", adapterID)
	if model != "" {
		fmt.Fprintf(&b, "Active model override: `%s`.\n", model)
	} else {
		b.WriteString("Active model: adapter default (no /model override set).\n")
	}
	b.WriteString("If the user asks which model or adapter is in use, answer with these values.\n\n")
	return b.String()
}

func truncateForContext(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func runShell(ctx context.Context, opts Options, line string) {
	if line == "" {
		fmt.Fprintln(opts.Out, "shell: empty")
		return
	}
	fmt.Fprintf(opts.Out, "  $ %s\n", line)
	c := exec.CommandContext(ctx, "sh", "-c", line)
	c.Dir = opts.Root
	c.Stdout = opts.Out
	c.Stderr = opts.Out
	if err := c.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ exit %v\n", err)
	}
}

func runHarnessCmd(ctx context.Context, opts Options, args []string) {
	fmt.Fprintf(opts.Out, "  $ harness %s\n", strings.Join(args, " "))
	c := exec.CommandContext(ctx, opts.HarnessBin, args...)
	c.Dir = opts.Root
	c.Stdout = opts.Out
	c.Stderr = opts.Out
	if err := c.Run(); err != nil {
		fmt.Fprintf(opts.Out, "  ✗ exit %v\n", err)
	}
}

type prefixWriter struct {
	w            io.Writer
	prefix       string
	buf          []byte
	atBOL        bool
	onFirstWrite func()
	firstWritten bool
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	if !p.firstWritten && len(b) > 0 {
		p.firstWritten = true
		if p.onFirstWrite != nil {
			p.onFirstWrite()
		}
	}
	if p.buf == nil {
		p.atBOL = true
	}
	for _, c := range b {
		if p.atBOL {
			p.buf = append(p.buf, []byte(p.prefix)...)
			p.atBOL = false
		}
		p.buf = append(p.buf, c)
		if c == '\n' {
			p.atBOL = true
		}
	}
	if p.atBOL {
		_, err := p.w.Write(p.buf)
		p.buf = p.buf[:0]
		if err != nil {
			return len(b), err
		}
	}
	return len(b), nil
}

func (p *prefixWriter) flush() {
	if len(p.buf) > 0 {
		_, _ = p.w.Write(p.buf)
		_, _ = p.w.Write([]byte{'\n'})
		p.buf = p.buf[:0]
	}
}
