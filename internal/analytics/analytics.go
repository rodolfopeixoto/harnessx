// SPDX-License-Identifier: MIT

// Package analytics aggregates chat session spend across one or many
// project roots so users can see how their token budget is being
// distributed by stack, adapter, task tag, and day.
package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/repl"
)

type Row struct {
	Stack     string
	Sessions  int
	Turns     int
	ChatTurns int
	InTokens  int
	OutTokens int
	CostUSD   float64
}

type AdapterRow struct {
	AdapterID string
	Task      string
	Turns     int
	CostUSD   float64
}

type DayRow struct {
	Day      string
	CostUSD  float64
	Turns    int
	Adapters int
}

type Report struct {
	Roots      []string
	Stacks     []Row
	Adapters   []AdapterRow
	Days       []DayRow
	TotalUSD   float64
	TotalTurns int
}

func Walk(roots []string, since time.Time) (Report, error) {
	r := Report{Roots: append([]string{}, roots...)}
	byStack := map[string]*Row{}
	byAdapter := map[string]*AdapterRow{}
	byDay := map[string]*DayRow{}

	for _, root := range roots {
		err := walkSessions(root, since, func(stack string, sess *repl.Session) {
			collectStackRow(byStack, stack, sess)
			for _, t := range sess.Turns {
				if t.Time.IsZero() {
					continue
				}
				if !since.IsZero() && t.Time.Before(since) {
					continue
				}
				collectAdapterRow(byAdapter, t)
				collectDayRow(byDay, t)
				if t.AdapterID != "" || t.CostUSD > 0 {
					r.TotalUSD += t.CostUSD
				}
				r.TotalTurns++
			}
		})
		if err != nil {
			return r, err
		}
	}

	for _, root := range roots {
		if err := walkRuns(root, since, func(stack string, m runMeta) {
			collectStackFromRun(byStack, stack, m)
			collectAdapterFromRun(byAdapter, m)
			collectDayFromRun(byDay, m)
			r.TotalUSD += m.CostUSD
			r.TotalTurns++
		}); err != nil {
			return r, err
		}
	}

	r.Stacks = sortStacks(byStack)
	r.Adapters = sortAdapters(byAdapter)
	r.Days = sortDays(byDay)
	return r, nil
}

// runMeta decouples analytics from internal/execution.Result so this
// package stays dependency-free of the executor.
type runMeta struct {
	RunID       string
	AgentID     string
	TaskTag     string
	StartedAt   time.Time
	CostUSD     float64
	InputToken  int
	OutputToken int
}

func walkRuns(root string, since time.Time, hit func(stack string, m runMeta)) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) || os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() != ".harness" {
			return nil
		}
		projectRoot := filepath.Dir(path)
		stack := detectStack(projectRoot)
		runsDir := filepath.Join(path, "runs")
		entries, rerr := os.ReadDir(runsDir)
		if rerr != nil {
			return filepath.SkipDir
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			meta, ok := loadRunMeta(filepath.Join(runsDir, e.Name(), "meta.json"))
			if !ok {
				continue
			}
			if !since.IsZero() && !meta.StartedAt.IsZero() && meta.StartedAt.Before(since) {
				continue
			}
			hit(stack, meta)
		}
		return filepath.SkipDir
	})
}

func loadRunMeta(path string) (runMeta, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return runMeta{}, false
	}
	var raw struct {
		RunID            string    `json:"run_id"`
		AgentID          string    `json:"agent_id"`
		TaskTag          string    `json:"task_tag"`
		StartedAt        time.Time `json:"started_at"`
		EstimatedCostUSD float64   `json:"estimated_cost_usd"`
		InputTokens      int       `json:"input_tokens"`
		OutputTokens     int       `json:"output_tokens"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return runMeta{}, false
	}
	return runMeta{
		RunID:       raw.RunID,
		AgentID:     raw.AgentID,
		TaskTag:     raw.TaskTag,
		StartedAt:   raw.StartedAt,
		CostUSD:     raw.EstimatedCostUSD,
		InputToken:  raw.InputTokens,
		OutputToken: raw.OutputTokens,
	}, true
}

func collectStackFromRun(byStack map[string]*Row, stack string, m runMeta) {
	row, ok := byStack[stack]
	if !ok {
		row = &Row{Stack: stack}
		byStack[stack] = row
	}
	row.Turns++
	row.InTokens += m.InputToken
	row.OutTokens += m.OutputToken
	row.CostUSD += m.CostUSD
}

func collectAdapterFromRun(by map[string]*AdapterRow, m runMeta) {
	id := m.AgentID
	if id == "" {
		id = "unknown"
	}
	key := id + "|" + m.TaskTag
	row, ok := by[key]
	if !ok {
		row = &AdapterRow{AdapterID: id, Task: m.TaskTag}
		by[key] = row
	}
	row.Turns++
	row.CostUSD += m.CostUSD
}

func collectDayFromRun(by map[string]*DayRow, m runMeta) {
	if m.StartedAt.IsZero() {
		return
	}
	day := m.StartedAt.UTC().Format("2006-01-02")
	row, ok := by[day]
	if !ok {
		row = &DayRow{Day: day}
		by[day] = row
	}
	row.Turns++
	row.CostUSD += m.CostUSD
}

//nolint:gocognit // filepath.WalkDir callback fans out across dir filters
func walkSessions(root string, since time.Time, hit func(stack string, sess *repl.Session)) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) || os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".harness" {
			projectRoot := filepath.Dir(path)
			stack := detectStack(projectRoot)
			ids, lerr := repl.ListSessions(projectRoot)
			if lerr != nil {
				return nil
			}
			for _, s := range ids {
				sess, lerr := repl.LoadSession(projectRoot, s.ID)
				if lerr != nil || sess == nil {
					continue
				}
				if !since.IsZero() && sess.Started.Before(since) {
					continue
				}
				hit(stack, sess)
			}
			return filepath.SkipDir
		}
		return nil
	})
}

func detectStack(root string) string {
	probes := []struct {
		marker string
		stack  string
	}{
		{"Cargo.toml", "rust"},
		{"go.mod", "go"},
		{"Gemfile", "ruby"},
		{"pyproject.toml", "python"},
		{"requirements.txt", "python"},
		{"package.json", "node"},
	}
	for _, p := range probes {
		if _, err := os.Stat(filepath.Join(root, p.marker)); err == nil {
			if p.stack == "ruby" {
				if _, err := os.Stat(filepath.Join(root, "config", "application.rb")); err == nil {
					return "rails"
				}
			}
			return p.stack
		}
	}
	return "unknown"
}

func collectStackRow(byStack map[string]*Row, stack string, sess *repl.Session) {
	row, ok := byStack[stack]
	if !ok {
		row = &Row{Stack: stack}
		byStack[stack] = row
	}
	row.Sessions++
	row.Turns += len(sess.Turns)
	for _, t := range sess.Turns {
		if t.Action == "chat" {
			row.ChatTurns++
		}
		row.InTokens += t.InTokens
		row.OutTokens += t.OutTokens
		row.CostUSD += t.CostUSD
	}
}

func collectAdapterRow(by map[string]*AdapterRow, t repl.Turn) {
	if t.AdapterID == "" && t.CostUSD == 0 {
		return
	}
	id := t.AdapterID
	if id == "" {
		id = "unknown"
	}
	key := id + "|" + t.TaskTag
	row, ok := by[key]
	if !ok {
		row = &AdapterRow{AdapterID: id, Task: t.TaskTag}
		by[key] = row
	}
	row.Turns++
	row.CostUSD += t.CostUSD
}

func collectDayRow(by map[string]*DayRow, t repl.Turn) {
	if t.Time.IsZero() {
		return
	}
	day := t.Time.UTC().Format("2006-01-02")
	row, ok := by[day]
	if !ok {
		row = &DayRow{Day: day}
		by[day] = row
	}
	row.Turns++
	row.CostUSD += t.CostUSD
}

func sortStacks(m map[string]*Row) []Row {
	out := make([]Row, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CostUSD > out[j].CostUSD })
	return out
}

func sortAdapters(m map[string]*AdapterRow) []AdapterRow {
	out := make([]AdapterRow, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CostUSD > out[j].CostUSD })
	return out
}

func sortDays(m map[string]*DayRow) []DayRow {
	out := make([]DayRow, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Day < out[j].Day })
	return out
}

func Render(out io.Writer, r Report) {
	fmt.Fprintln(out, "harness analytics")
	if len(r.Roots) > 0 {
		fmt.Fprintf(out, "  roots: %s\n", strings.Join(r.Roots, ", "))
	}
	fmt.Fprintf(out, "  total turns: %d   total cost: $%.4f\n\n", r.TotalTurns, r.TotalUSD)

	fmt.Fprintln(out, "by stack")
	fmt.Fprintf(out, "  %-10s %5s %5s %5s %10s %10s %12s\n",
		"STACK", "SESS", "TURN", "CHAT", "IN", "OUT", "COST")
	for _, s := range r.Stacks {
		fmt.Fprintf(out, "  %-10s %5d %5d %5d %10d %10d $%11.4f\n",
			s.Stack, s.Sessions, s.Turns, s.ChatTurns, s.InTokens, s.OutTokens, s.CostUSD)
	}

	if len(r.Adapters) > 0 {
		fmt.Fprintln(out, "\nby adapter / task")
		fmt.Fprintf(out, "  %-12s %-16s %5s %12s\n", "ADAPTER", "TASK", "TURNS", "COST")
		for _, a := range r.Adapters {
			task := a.Task
			if task == "" {
				task = "-"
			}
			fmt.Fprintf(out, "  %-12s %-16s %5d $%11.4f\n", a.AdapterID, task, a.Turns, a.CostUSD)
		}
	}

	if len(r.Days) > 0 {
		fmt.Fprintln(out, "\nby day")
		fmt.Fprintf(out, "  %-10s %5s %12s\n", "DAY", "TURNS", "COST")
		for _, d := range r.Days {
			fmt.Fprintf(out, "  %-10s %5d $%11.4f\n", d.Day, d.Turns, d.CostUSD)
		}
	}
}

func RenderJSON(out io.Writer, r Report) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
