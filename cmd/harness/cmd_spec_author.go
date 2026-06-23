// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/activeagent"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/specflow"
	"github.com/ropeixoto/harnessx/internal/ui"
)

func newSpecAuthorCmd() *cobra.Command {
	var (
		adapterID   string
		nonInter    bool
		skipAsk     bool
		acceptDraft bool
	)
	c := &cobra.Command{
		Use:   "author \"<feature prompt>\"",
		Short: "Interactive spec authoring loop (clarify → draft → refine → save)",
		Long: `Walks the user through an editable spec:
1. ask baseline + LLM-derived clarifying questions
2. draft markdown via the planning chain
3. edit-loop: /edit (in $EDITOR), /refine "...", /expand <sec>, /shrink <sec>,
   /diff, /undo, /show, /save, /cancel
4. persist to .harness/artifacts/specs/<id>.md + <id>.history.jsonl

Adapter follows .harness/config/active.yaml unless --adapter is set.
Use --no-interactive to write a baseline-only spec without prompting.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := cwd()
			if err != nil {
				return err
			}
			prompt := strings.Join(args, " ")
			out := cmd.OutOrStdout()
			in := cmd.InOrStdin()

			adapter := resolveSpecAuthorAdapter(root, adapterID, out)
			sess := specflow.New(root, prompt)

			if nonInter {
				if _, err := sess.Draft(cmd.Context(), adapter); err != nil {
					return err
				}
				p, err := sess.Save()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "spec saved → %s\n", p)
				return nil
			}

			reader := bufio.NewReader(in)
			if !skipAsk {
				sess.Questions = append(sess.Questions, specflow.BaselineQuestions...)
				sess.Questions = append(sess.Questions, specflow.ContextQuestions(cmd.Context(), adapter, prompt)...)
				if err := collectSpecAnswers(reader, out, sess); err != nil {
					return err
				}
			}

			fmt.Fprintln(out, ui.Muted.Render("drafting spec…"))
			if _, err := sess.Draft(cmd.Context(), adapter); err != nil {
				return err
			}
			fmt.Fprintln(out, sess.Body)

			if acceptDraft {
				p, err := sess.Save()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "spec saved → %s\n", p)
				return nil
			}

			return runSpecAuthorREPL(cmd.Context(), reader, out, sess, adapter)
		},
	}
	c.Flags().StringVar(&adapterID, "adapter", "", "adapter id (default: active pin, then first registered)")
	c.Flags().BoolVar(&nonInter, "no-interactive", false, "skip Q&A + edit loop; emit baseline spec only")
	c.Flags().BoolVar(&skipAsk, "skip-questions", false, "skip clarifying questions; jump to draft")
	c.Flags().BoolVar(&acceptDraft, "accept-draft", false, "save the first draft without entering the edit loop")
	return c
}

func resolveSpecAuthorAdapter(root, override string, out io.Writer) agents.AgentAdapter {
	id := activeagent.ResolveAgentID(root, override)
	if id == "" {
		return nil
	}
	reg, _, err := agentcmd.LoadAll(root)
	if err != nil {
		return nil
	}
	a, ok := reg.Get(id)
	if !ok {
		fmt.Fprintf(out, "spec: adapter %q not registered — running offline (baseline spec only)\n", id)
		return nil
	}
	return a
}

func collectSpecAnswers(r *bufio.Reader, out io.Writer, sess *specflow.Session) error {
	fmt.Fprintln(out, "\n"+ui.Heading.Render("clarifying questions")+" — empty line skips optional ones")
	for _, q := range sess.Questions {
		tag := "  "
		if q.Required {
			tag = ui.Accent.Render("*") + " "
		}
		fmt.Fprintf(out, "%s%s: ", tag, q.Prompt)
		line, _ := r.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if q.Required && strings.TrimSpace(line) == "" {
			fmt.Fprintln(out, "  "+ui.MarkWarn()+" required — please answer:")
			again, _ := r.ReadString('\n')
			line = strings.TrimRight(again, "\r\n")
		}
		sess.Answers = append(sess.Answers, specflow.Answer{Key: q.Key, Value: line})
	}
	return nil
}

func runSpecAuthorREPL(ctx context.Context, r *bufio.Reader, out io.Writer, sess *specflow.Session, adapter agents.AgentAdapter) error {
	fmt.Fprintln(out, "\n"+ui.Heading.Render("spec edit loop")+" — /help for commands")
	for {
		fmt.Fprint(out, ui.Accent.Render("spec> "))
		line, err := r.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		switch {
		case line == "" || line == "/help":
			specAuthorHelp(out)
		case line == "/show":
			sess.Render(out)
		case line == "/sections":
			for i, s := range sess.SectionList() {
				fmt.Fprintf(out, "  %d. %s\n", i+1, s)
			}
		case line == "/diff":
			d := sess.Diff()
			if d == "" {
				fmt.Fprintln(out, "  (no prior revision)")
			} else {
				fmt.Fprintln(out, d)
			}
		case line == "/undo":
			if _, err := sess.Undo(); err != nil {
				fmt.Fprintf(out, "  %s %v\n", ui.MarkWarn(), err)
				continue
			}
			fmt.Fprintln(out, "  ✓ reverted one revision")
		case line == "/edit":
			edited, err := specflow.EditViaEditor(sess.Body)
			if err != nil {
				fmt.Fprintf(out, "  %s editor: %v\n", ui.MarkFail(), err)
				continue
			}
			sess.ApplyEdit(edited)
			fmt.Fprintln(out, "  ✓ edit applied")
		case strings.HasPrefix(line, "/refine "):
			if err := specRefine(ctx, out, sess, adapter, strings.TrimPrefix(line, "/refine ")); err != nil {
				fmt.Fprintf(out, "  %s %v\n", ui.MarkFail(), err)
			}
		case strings.HasPrefix(line, "/expand"):
			section := strings.TrimSpace(strings.TrimPrefix(line, "/expand"))
			if _, err := sess.Expand(ctx, adapter, section); err != nil {
				fmt.Fprintf(out, "  %s %v\n", ui.MarkFail(), err)
				continue
			}
			fmt.Fprintln(out, "  ✓ expanded — /diff to inspect")
		case strings.HasPrefix(line, "/shrink"):
			section := strings.TrimSpace(strings.TrimPrefix(line, "/shrink"))
			if _, err := sess.Shrink(ctx, adapter, section); err != nil {
				fmt.Fprintf(out, "  %s %v\n", ui.MarkFail(), err)
				continue
			}
			fmt.Fprintln(out, "  ✓ tightened — /diff to inspect")
		case line == "/save":
			p, err := sess.Save()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "  %s saved → %s\n", ui.MarkSuccess(), p)
			return nil
		case line == "/cancel":
			fmt.Fprintln(out, "  cancelled — nothing written")
			return nil
		default:
			fmt.Fprintln(out, "  unknown command — /help for the list")
		}
	}
}

func specRefine(ctx context.Context, out io.Writer, sess *specflow.Session, adapter agents.AgentAdapter, rest string) error {
	section := ""
	instr := rest
	if i := strings.Index(rest, ":"); i > 0 && i < 40 && !strings.HasPrefix(rest, "\"") {
		section = strings.TrimSpace(rest[:i])
		instr = strings.TrimSpace(rest[i+1:])
	}
	if strings.TrimSpace(instr) == "" {
		return fmt.Errorf("usage: /refine [section:] <instruction>")
	}
	if _, err := sess.Refine(ctx, adapter, section, instr); err != nil {
		return err
	}
	fmt.Fprintln(out, "  ✓ refined — /diff to inspect, /undo to revert")
	return nil
}

func specAuthorHelp(w io.Writer) {
	for _, ln := range []string{
		"  /show                  — print current spec",
		"  /sections              — list H2 sections",
		"  /edit                  — open $EDITOR on the draft",
		"  /refine [sec:] <instr> — LLM rewrite (sec optional)",
		"  /expand [sec]          — LLM adds detail",
		"  /shrink [sec]          — LLM tightens",
		"  /diff                  — last revision delta",
		"  /undo                  — drop the last revision",
		"  /save                  — write .harness/artifacts/specs/<id>.md + history",
		"  /cancel                — discard and exit",
	} {
		fmt.Fprintln(w, ln)
	}
}
