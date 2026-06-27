package runscmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/ropeixoto/harnessx/internal/execution"
)

type Options struct {
	Root  string
	JSON  bool
	RunID string
}

func List(out io.Writer, opts Options) error {
	runs, err := execution.ListRuns(opts.Root)
	if err != nil {
		return err
	}
	if opts.JSON {
		return json.NewEncoder(out).Encode(runs)
	}
	return renderTable(out, runs)
}

func Inspect(out io.Writer, opts Options) error {
	r, err := execution.LoadRun(opts.Root, opts.RunID)
	if err != nil {
		if errors.Is(err, execution.ErrRunIncomplete) {
			return fmt.Errorf("run %s exists on disk but has no meta.json; check .harness/runs/%s/report.md or run `harness runs prune` to clean orphans", opts.RunID, opts.RunID)
		}
		return err
	}
	if opts.JSON {
		return json.NewEncoder(out).Encode(r)
	}
	return renderDetail(out, r)
}

func Report(out io.Writer, opts Options) error {
	reportPath := ""
	r, err := execution.LoadRun(opts.Root, opts.RunID)
	if err != nil {
		if !errors.Is(err, execution.ErrRunIncomplete) {
			return err
		}
		reportPath = filepath.Join(opts.Root, ".harness", "runs", opts.RunID, "report.md")
	} else {
		reportPath = r.ReportPath
	}
	if reportPath == "" {
		return fmt.Errorf("run %s has no report.md", opts.RunID)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

func Sensors(out io.Writer, opts Options) error {
	r, err := execution.LoadRun(opts.Root, opts.RunID)
	if err != nil {
		if errors.Is(err, execution.ErrRunIncomplete) {
			return fmt.Errorf("run %s incomplete (no meta.json); sensor outcomes are only persisted for runs created by `harness run/auto/ship`", opts.RunID)
		}
		return err
	}
	if len(r.Sensors) == 0 {
		fmt.Fprintln(out, "no sensors recorded")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tMS\tOUTPUT")
	for _, s := range r.Sensors {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.ID, s.Status, s.DurationMs, s.Output)
	}
	return w.Flush()
}

func renderTable(out io.Writer, runs []execution.Result) error {
	if len(runs) == 0 {
		fmt.Fprintln(out, "no runs")
		return nil
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "RUN ID\tAGENT\tSTATUS\tFILES\tCOST")
	for _, r := range runs {
		agent := r.AgentID
		if agent == "" {
			agent = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t$%.4f\n",
			r.RunID, agent, r.Status, len(r.ChangedFiles), r.EstimatedCostUSD)
	}
	return w.Flush()
}

func renderDetail(out io.Writer, r execution.Result) error {
	fmt.Fprintf(out, "Run:        %s\nAgent:      %s\nStatus:     %s\nStarted:    %s\nFinished:   %s\nWorktree:   %s\nStdout:     %s\nStderr:     %s\nDiff:       %s\nReport:     %s\nFiles:      %d\nTokens:     in=%d out=%d\nCost (est): $%.4f (exact=%t)\n",
		r.RunID, r.AgentID, r.Status, r.StartedAt.Format("2006-01-02 15:04:05"), r.FinishedAt.Format("2006-01-02 15:04:05"),
		r.WorktreePath, r.StdoutPath, r.StderrPath, r.DiffPath, r.ReportPath,
		len(r.ChangedFiles), r.InputTokens, r.OutputTokens, r.EstimatedCostUSD, r.ExactUsageAvailable)
	if len(r.Sensors) > 0 {
		fmt.Fprintln(out, "Sensors:")
		for _, s := range r.Sensors {
			fmt.Fprintf(out, "  - %s [%s] %dms\n", s.ID, s.Status, s.DurationMs)
		}
	}
	if len(r.MCPDetectedNotActive) > 0 {
		fmt.Fprintf(out, "\nMCP configs detected but not injected yet (P32): %d\n", len(r.MCPDetectedNotActive))
	}
	if len(r.HooksDetectedNotActive) > 0 {
		fmt.Fprintf(out, "Hooks detected but not executed yet (P32): %d\n", len(r.HooksDetectedNotActive))
	}
	if r.ErrorMessage != "" {
		fmt.Fprintf(out, "\nError: %s: %s\n", r.ErrorType, r.ErrorMessage)
	}
	return nil
}
