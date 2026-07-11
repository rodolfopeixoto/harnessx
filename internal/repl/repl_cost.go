// SPDX-License-Identifier: MIT

package repl

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// adapterBillingMode tells the user whether the adapter is going to
// charge against their API key (oneshot CLI: claude --print, codex
// exec, gemini -p) or against a logged-in plan/subscription
// (interactive CLI: claude-interactive, kimi chat). The distinction
// matters because oneshot calls show up on a separate invoice while
// interactive ones drain the user's chat-mode token quota.
func adapterBillingMode(id string) string {
	switch id {
	case "claude", "codex", "gemini", "anthropic-api", "openai-api",
		"gemini-api", "moonshot-api", "minimax-api":
		return "oneshot · API-billed"
	case "claude-interactive", "kimi", "ollama":
		return "interactive · plan/local"
	case "fake":
		return "fake · free"
	}
	return "unknown billing"
}

func printAgents(opts Options) {
	if len(opts.AdaptersList) == 0 {
		fmt.Fprintln(opts.Out, "no adapters registered (run 'harness agent list' for details)")
		return
	}
	fmt.Fprintf(opts.Out, "active: %s\n", opts.AdapterID)
	for _, id := range opts.AdaptersList {
		marker := "  "
		if id == opts.AdapterID {
			marker = "→ "
		}
		fmt.Fprintf(opts.Out, "%s%s\n", marker, id)
	}
}

func printCost(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "no agent turns yet")
		return
	}
	totals := aggregateCost(sess.Turns)
	renderCostReport(out, sess.ID, totals)
}

type costRow struct {
	AdapterID string
	Task      string
	Turns     int
	InTokens  int
	OutTokens int
	CostUSD   float64
}

type costTotals struct {
	ChatTurns int
	Total     costRow
	PerAgent  []costRow
}

func aggregateCost(turns []Turn) costTotals {
	byKey := map[string]*costRow{}
	keys := []string{}
	total := costRow{AdapterID: "TOTAL"}
	chatTurns := 0
	for _, t := range turns {
		if t.Action == "chat" {
			chatTurns++
		}
		if t.InTokens == 0 && t.OutTokens == 0 && t.CostUSD == 0 {
			continue
		}
		id := t.AdapterID
		if id == "" {
			id = "unknown"
		}
		key := id + "|" + t.TaskTag
		row, ok := byKey[key]
		if !ok {
			row = &costRow{AdapterID: id, Task: t.TaskTag}
			byKey[key] = row
			keys = append(keys, key)
		}
		row.Turns++
		row.InTokens += t.InTokens
		row.OutTokens += t.OutTokens
		row.CostUSD += t.CostUSD
		total.Turns++
		total.InTokens += t.InTokens
		total.OutTokens += t.OutTokens
		total.CostUSD += t.CostUSD
	}
	rows := make([]costRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, *byKey[k])
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].CostUSD > rows[j].CostUSD })
	return costTotals{ChatTurns: chatTurns, Total: total, PerAgent: rows}
}

func renderCostReport(out io.Writer, sessionID string, t costTotals) {
	fmt.Fprintf(out, "session %s: %d chat turns\n", sessionID, t.ChatTurns)
	if len(t.PerAgent) == 0 {
		fmt.Fprintln(out, "  (no recorded usage)")
		return
	}
	fmt.Fprintf(out, "  %-12s %-16s %5s %8s %8s %10s\n", "ADAPTER", "TASK", "TURNS", "IN", "OUT", "COST")
	for _, r := range t.PerAgent {
		task := r.Task
		if task == "" {
			task = "-"
		}
		fmt.Fprintf(out, "  %-12s %-16s %5d %8d %8d $%9.4f\n",
			r.AdapterID, task, r.Turns, r.InTokens, r.OutTokens, r.CostUSD)
	}
	fmt.Fprintln(out, "  "+strings.Repeat("─", 64))
	fmt.Fprintf(out, "  %-12s %-16s %5d %8d %8d $%9.4f\n",
		"TOTAL", "", t.Total.Turns, t.Total.InTokens, t.Total.OutTokens, t.Total.CostUSD)
}

// printTimeline renders an at-a-glance ASCII view of every turn in
// the session: clock, action label, truncated input, and cost in
// USD. Designed for the "what happened today?" lookback after a long
// chat. Cumulative cost is printed at the foot for a one-line sanity
// check.
func printTimeline(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "  no turns yet")
		return
	}
	var total float64
	for i, t := range sess.Turns {
		clock := t.Time.Local().Format("15:04:05")
		input := truncateForContext(t.Input, 60)
		cost := ""
		if t.CostUSD > 0 {
			cost = fmt.Sprintf("  $%.4f", t.CostUSD)
			total += t.CostUSD
		}
		fmt.Fprintf(out, "  %3d  %s  [%-15s] %s%s\n", i+1, clock, t.Action, input, cost)
	}
	fmt.Fprintf(out, "\n  total: %d turns, ~$%.4f\n", len(sess.Turns), total)
}
