// SPDX-License-Identifier: MIT

package repl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/intentplan"
	"github.com/ropeixoto/harnessx/internal/ui"
)

type SessionSummary struct {
	ID        string
	Label     string
	Goal      intentplan.Goal
	Turns     int
	LastInput string
}

// LoadSession rehydrates a prior chat from .harness/sessions/<id>.jsonl.
// The file is JSONL of Turn records (see persist) without the Session
// envelope, so we synthesise a Session shell and replay every turn into
// it. Used by `harness chat --resume <id>`.
func LoadSession(root, id string) (*Session, error) {
	p := sessionPath(root, id)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sess := Session{ID: id, Started: time.Now().UTC(), Root: root}
	dec := json.NewDecoder(f)
	for dec.More() {
		var t Turn
		if err := dec.Decode(&t); err != nil {
			return nil, err
		}
		sess.Turns = append(sess.Turns, t)
	}
	if meta, err := loadSessionMeta(root, id); err == nil {
		if meta.Goal != "" {
			sess.Goal = intentplan.Goal(meta.Goal)
		}
		sess.Label = meta.Label
		sess.ContextMark = meta.ContextMark
		sess.AutoGate = meta.AutoGate
		sess.BudgetUSD = meta.BudgetUSD
	}
	if sess.Goal == "" {
		for i := len(sess.Turns) - 1; i >= 0; i-- {
			if sess.Turns[i].Plan != nil {
				sess.Goal = sess.Turns[i].Plan.Goal
				break
			}
		}
	}
	if sess.Goal == "" {
		sess.Goal = intentplan.GoalDev
	}
	return &sess, nil
}

// ResolveSessionID converts either a ulid or a /save label into the
// canonical session id. Labels are matched against ListSessions; on
// ambiguity (two sessions sharing a label) the newest wins. Returns
// the input unchanged when no label matches so callers can still
// load by raw ulid.
func ResolveSessionID(root, arg string) string {
	if arg == "" {
		return arg
	}
	rows, err := ListSessions(root)
	if err != nil {
		return arg
	}
	for _, r := range rows {
		if r.ID == arg {
			return arg
		}
	}
	for _, r := range rows {
		if r.Label == arg {
			return r.ID
		}
	}
	return arg
}

// SuggestSession returns the closest known label or id to arg via
// Levenshtein distance, with a max distance of 3. Empty when nothing
// is close enough — callers use the empty case to fall through to
// the canonical "session not found" error so we never auto-resolve
// to the wrong session silently.
func SuggestSession(root, arg string) string {
	if arg == "" {
		return ""
	}
	rows, err := ListSessions(root)
	if err != nil || len(rows) == 0 {
		return ""
	}
	candidates := make([]string, 0, 2*len(rows))
	for _, r := range rows {
		if r.Label != "" {
			candidates = append(candidates, r.Label)
		}
		candidates = append(candidates, r.ID)
	}
	best := ""
	bestDist := 4
	for _, c := range candidates {
		d := levenshtein(arg, c)
		if d < bestDist {
			best = c
			bestDist = d
		}
	}
	return best
}

// ListSessions returns one summary per .harness/sessions/*.jsonl file,
// sorted newest first by file mtime.
//
//nolint:gocognit // single WalkDir pass with several inline filters
func ListSessions(root string) ([]SessionSummary, error) {
	dir := filepath.Join(root, ".harness", "sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type entryWithStat struct {
		name  string
		mtime time.Time
	}
	var stats []entryWithStat
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		stats = append(stats, entryWithStat{name: e.Name(), mtime: info.ModTime()})
	}
	for i := range stats {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].mtime.After(stats[i].mtime) {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	out := make([]SessionSummary, 0, len(stats))
	for _, s := range stats {
		id := strings.TrimSuffix(s.name, ".jsonl")
		sess, err := LoadSession(root, id)
		if err != nil {
			continue
		}
		last := ""
		for i := len(sess.Turns) - 1; i >= 0; i-- {
			if sess.Turns[i].Input != "" {
				last = sess.Turns[i].Input
				if len(last) > 60 {
					last = last[:60] + "…"
				}
				break
			}
		}
		out = append(out, SessionSummary{
			ID: sess.ID, Label: sess.Label,
			Goal: sess.Goal, Turns: len(sess.Turns), LastInput: last,
		})
	}
	return out, nil
}

// setSessionLabel tags the session with a human-readable alias so
// `harness chat list` and exports show "shop-api-checkout" instead of
// an opaque ulid. Refuses obviously broken inputs (slashes, dot
// prefix) so it stays safe to interpolate into paths and CLI output.
func setSessionLabel(sess *Session, out io.Writer, label string) {
	if label == "" {
		fmt.Fprintln(out, "  ✗ /save needs a name (try /save my-feature)")
		return
	}
	if strings.ContainsAny(label, "/\\\n\t") || strings.HasPrefix(label, ".") {
		fmt.Fprintf(out, "  ✗ /save: %q contains an unsupported character\n", label)
		return
	}
	if len(label) > 80 {
		label = label[:80]
	}
	sess.Label = label
	fmt.Fprintf(out, "  ✓ session labelled %q\n", label)
}

// checkBudget returns false when running another chat turn would push
// cumulative spend over the configured cap. Prints why and refuses.
func checkBudget(sess *Session, out io.Writer) bool {
	if sess == nil || sess.BudgetUSD <= 0 {
		return true
	}
	var spent float64
	for _, t := range sess.Turns {
		spent += t.CostUSD
	}
	if spent >= sess.BudgetUSD {
		fmt.Fprintf(out, "  ✗ budget exhausted: $%.4f spent of $%.4f cap. /budget off to reset.\n",
			spent, sess.BudgetUSD)
		return false
	}
	return true
}

// setBudget parses a USD value from /budget and stores it on the
// session. Cumulative spend is enforced by checkBudget before each
// chat turn.
func setBudget(sess *Session, out io.Writer, raw string) {
	raw = strings.TrimPrefix(raw, "$")
	if raw == "" || raw == "off" || raw == "0" {
		sess.BudgetUSD = 0
		fmt.Fprintln(out, "  ✓ budget cleared")
		return
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		fmt.Fprintf(out, "  ✗ /budget needs a non-negative USD number (e.g. /budget 0.50)\n")
		return
	}
	sess.BudgetUSD = v
	fmt.Fprintf(out, "  ✓ budget set to $%.4f for this session\n", v)
}

func greet(out io.Writer, s Session) {
	fmt.Fprintf(out, "harness chat — session %s, goal=%s\n", s.ID, s.Goal)
	fmt.Fprintln(out, `plain text → talk to agent · /exec → plan+run · !<cmd> → shell · /help · /exit`)
	fmt.Fprintln(out, ui.Muted.Render(`multi-line: end line with \  ·  or wrap with """ … """  ·  / lists slashes`))
	fmt.Fprintln(out)
}

func summariseSession(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		return
	}
	totals := aggregateCost(sess.Turns)
	fmt.Fprintln(out, ui.Heading.Render("session recap"))
	fmt.Fprintf(out, "  %s id     %s\n", ui.Muted.Render("·"), ui.Accent.Render(sess.ID))
	if sess.Label != "" {
		fmt.Fprintf(out, "  %s label  %s\n", ui.Muted.Render("·"), ui.Accent.Render(sess.Label))
	}
	fmt.Fprintf(out, "  %s goal   %s\n", ui.Muted.Render("·"), string(sess.Goal))
	fmt.Fprintf(out, "  %s turns  %d (chat=%d)\n", ui.Muted.Render("·"), len(sess.Turns), totals.ChatTurns)
	fmt.Fprintf(out, "  %s tokens in=%d out=%d\n", ui.Muted.Render("·"), totals.Total.InTokens, totals.Total.OutTokens)
	fmt.Fprintf(out, "  %s cost   %s\n", ui.Muted.Render("·"),
		ui.Success.Render(fmt.Sprintf("$%.4f", totals.Total.CostUSD)))
}

func lastPromptInput(sess *Session) string {
	if sess == nil {
		return ""
	}
	for i := len(sess.Turns) - 1; i >= 0; i-- {
		in := sess.Turns[i].Input
		if in == "" || strings.HasPrefix(in, "/") {
			continue
		}
		return in
	}
	return ""
}

func printHistory(out io.Writer, sess *Session) {
	if sess == nil || len(sess.Turns) == 0 {
		fmt.Fprintln(out, "history empty")
		return
	}
	start := 0
	if len(sess.Turns) > 20 {
		start = len(sess.Turns) - 20
	}
	for i := start; i < len(sess.Turns); i++ {
		fmt.Fprintf(out, "%3d  %s\n", i+1, sess.Turns[i].Input)
	}
}

func sessionPath(root, id string) string {
	return filepath.Join(root, ".harness", "sessions", id+".jsonl")
}

func sessionMetaPath(root, id string) string {
	return filepath.Join(root, ".harness", "sessions", id+".meta.json")
}

// sessionMeta carries the Session fields that do not fit in the
// per-turn JSONL stream. Persisted as a sidecar alongside the JSONL so
// older readers (which ignore it) keep working.
type sessionMeta struct {
	ID          string  `json:"id"`
	Goal        string  `json:"goal"`
	Label       string  `json:"label,omitempty"`
	ContextMark int     `json:"context_mark,omitempty"`
	AutoGate    bool    `json:"auto_gate,omitempty"`
	BudgetUSD   float64 `json:"budget_usd,omitempty"`
}

func emitTurnJSON(w io.Writer, sessionID string, t Turn) {
	envelope := struct {
		Session   string    `json:"session"`
		Time      time.Time `json:"time"`
		Input     string    `json:"input"`
		Action    string    `json:"action"`
		Adapter   string    `json:"adapter_id,omitempty"`
		TaskTag   string    `json:"task_tag,omitempty"`
		InTokens  int       `json:"in_tokens,omitempty"`
		OutTokens int       `json:"out_tokens,omitempty"`
		CostUSD   float64   `json:"cost_usd,omitempty"`
		Ok        *bool     `json:"ok,omitempty"`
	}{
		Session:   sessionID,
		Time:      t.Time,
		Input:     t.Input,
		Action:    t.Action,
		Adapter:   t.AdapterID,
		TaskTag:   t.TaskTag,
		InTokens:  t.InTokens,
		OutTokens: t.OutTokens,
		CostUSD:   t.CostUSD,
	}
	if t.Result != nil {
		ok := t.Result.OK
		envelope.Ok = &ok
	}
	b, err := json.Marshal(envelope)
	if err != nil {
		return
	}
	_, _ = w.Write(append(b, '\n'))
}

func persist(root string, s Session) error {
	p := sessionPath(root, s.ID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, t := range s.Turns {
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	meta := sessionMeta{
		ID:          s.ID,
		Goal:        string(s.Goal),
		Label:       s.Label,
		ContextMark: s.ContextMark,
		AutoGate:    s.AutoGate,
		BudgetUSD:   s.BudgetUSD,
	}
	body, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionMetaPath(root, s.ID), body, 0o644)
}

func loadSessionMeta(root, id string) (sessionMeta, error) {
	body, err := os.ReadFile(sessionMetaPath(root, id))
	if err != nil {
		return sessionMeta{}, err
	}
	var m sessionMeta
	if err := json.Unmarshal(body, &m); err != nil {
		return sessionMeta{}, err
	}
	return m, nil
}
