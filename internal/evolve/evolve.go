// SPDX-License-Identifier: MIT

// Package evolve implements paper §3.5 Agentic Harness Engineering:
// deep-telemetry-driven harness mutation under governance. Diagnose
// scans .harness/logs/events.jsonl + run reports to surface failure
// clusters. Replay re-runs a candidate mutation against a held-out
// trace set. Promote requires HITL approval before writing the
// mutation record.
package evolve

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/ids"
)

type Event struct {
	Time   string         `json:"time"`
	Level  string         `json:"level"`
	Fields map[string]any `json:"fields"`
}

type FailureCluster struct {
	Signature string `json:"signature"`
	Count     int    `json:"count"`
	Example   string `json:"example"`
}

type Diagnosis struct {
	Generated time.Time        `json:"generated"`
	Events    int              `json:"events_scanned"`
	Failures  int              `json:"failures"`
	Clusters  []FailureCluster `json:"clusters"`
}

func eventsPath(root string) string {
	return filepath.Join(root, ".harness", "logs", "events.jsonl")
}

func Diagnose(root string) (Diagnosis, error) {
	f, err := os.Open(eventsPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return Diagnosis{Generated: time.Now().UTC()}, nil
	}
	if err != nil {
		return Diagnosis{}, err
	}
	defer f.Close()
	d := Diagnosis{Generated: time.Now().UTC()}
	clusters := map[string]*FailureCluster{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		d.Events++
		line := sc.Bytes()
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Level != "error" && !isFailureFields(e.Fields) {
			continue
		}
		d.Failures++
		sig := signature(e.Fields)
		c, ok := clusters[sig]
		if !ok {
			c = &FailureCluster{Signature: sig, Example: truncate(string(line), 256)}
			clusters[sig] = c
		}
		c.Count++
	}
	for _, c := range clusters {
		d.Clusters = append(d.Clusters, *c)
	}
	sort.Slice(d.Clusters, func(i, j int) bool { return d.Clusters[i].Count > d.Clusters[j].Count })
	if err := sc.Err(); err != nil {
		return d, err
	}
	return d, nil
}

func isFailureFields(m map[string]any) bool {
	if v, ok := m["status"].(string); ok {
		switch strings.ToLower(v) {
		case "failed", "fail", "error", "red", "denied":
			return true
		}
	}
	if v, ok := m["err"].(string); ok && v != "" {
		return true
	}
	if v, ok := m["error"].(string); ok && v != "" {
		return true
	}
	return false
}

func signature(m map[string]any) string {
	keys := []string{"stage", "sensor", "agent", "task", "status"}
	parts := []string{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	if len(parts) == 0 {
		return "uncategorised"
	}
	return strings.Join(parts, "|")
}

type Mutation struct {
	ID          string    `json:"id"`
	Component   string    `json:"component"`
	Description string    `json:"description"`
	Rationale   string    `json:"rationale"`
	Risk        string    `json:"risk"`
	HITL        bool      `json:"hitl_approved"`
	CreatedAt   time.Time `json:"created_at"`
}

func mutationsLogPath(root string) string {
	return filepath.Join(root, ".harness", "logs", "mutations.jsonl")
}

// Propose stages a mutation candidate. Status is "proposed" until
// Promote writes a HITL=true entry on top.
func Propose(root string, m Mutation) (string, error) {
	m.ID = ids.New()
	m.CreatedAt = time.Now().UTC()
	return m.ID, appendMutation(root, m, "proposed")
}

type PromoteOptions struct {
	MutationID string
	HITL       bool
	Reason     string
}

func Promote(root string, opts PromoteOptions) error {
	if !opts.HITL {
		return errors.New("evolve: promote requires --hitl (paper §3.5.3 governed mutation)")
	}
	rec := Mutation{ID: opts.MutationID, HITL: true, Rationale: opts.Reason, CreatedAt: time.Now().UTC()}
	return appendMutation(root, rec, "promoted")
}

func appendMutation(root string, m Mutation, status string) error {
	if err := os.MkdirAll(filepath.Dir(mutationsLogPath(root)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(mutationsLogPath(root), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	out := map[string]any{
		"status":   status,
		"mutation": m,
	}
	enc := json.NewEncoder(f)
	return enc.Encode(out)
}

type ReplayResult struct {
	Replayed int    `json:"replayed"`
	Matched  int    `json:"matched"`
	Source   string `json:"source"`
}

// Replay walks the candidate trace JSONL and counts entries whose
// signature matches at least one cluster in the diagnosis. A real
// Evolution Agent would re-execute the trace against a mutated harness
// in a sandbox; this MVP scores text-level signature matches as
// regression proxy.
func Replay(root, traceFile string, d Diagnosis) (ReplayResult, error) {
	f, err := os.Open(traceFile)
	if err != nil {
		return ReplayResult{}, err
	}
	defer f.Close()
	known := map[string]bool{}
	for _, c := range d.Clusters {
		known[c.Signature] = true
	}
	res := ReplayResult{Source: traceFile}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		res.Replayed++
		var e Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		if known[signature(e.Fields)] {
			res.Matched++
		}
	}
	return res, sc.Err()
}

func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
