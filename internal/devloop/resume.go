// SPDX-License-Identifier: MIT

package devloop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

const StateSchemaVersion = 1

type State struct {
	SchemaVersion      int       `json:"schema_version"`
	RunID              string    `json:"run_id"`
	OriginalPrompt     string    `json:"original_prompt"`
	BaselineLintOK     bool      `json:"baseline_lint_ok"`
	BaselineTestOK     bool      `json:"baseline_test_ok"`
	Attempts           []Attempt `json:"attempts"`
	BudgetUSDRemaining float64   `json:"budget_usd_remaining"`
	MaxAttempts        int       `json:"max_attempts"`
	AgentID            string    `json:"agent_id"`
	Autonomy           string    `json:"autonomy"`
	Apply              bool      `json:"apply"`
	LintCmd            string    `json:"lint_cmd"`
	TestCmd            string    `json:"test_cmd"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func StateDir(root, runID string) string {
	return filepath.Join(paths.HarnessDir(root), "runs", "_loop", runID)
}

func StatePath(root, runID string) string {
	return filepath.Join(StateDir(root, runID), "state.json")
}

func WriteState(root string, s State) error {
	if s.RunID == "" {
		return fmt.Errorf("devloop: WriteState requires RunID")
	}
	s.SchemaVersion = StateSchemaVersion
	s.UpdatedAt = time.Now().UTC()
	dir := StateDir(root, s.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(root, s.RunID), body, 0o644)
}

func LoadState(root, runID string) (State, error) {
	body, err := os.ReadFile(StatePath(root, runID))
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(body, &s); err != nil {
		return State{}, fmt.Errorf("devloop: parse state.json: %w", err)
	}
	if s.SchemaVersion != StateSchemaVersion {
		return State{}, fmt.Errorf("devloop: unsupported state schema_version=%d (want %d)", s.SchemaVersion, StateSchemaVersion)
	}
	return s, nil
}

type ResumableRun struct {
	RunID     string
	UpdatedAt time.Time
	Attempts  int
	Remaining float64
	Prompt    string
}

func ListResumable(root string) ([]ResumableRun, error) {
	base := filepath.Join(paths.HarnessDir(root), "runs", "_loop")
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []ResumableRun
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		s, err := LoadState(root, e.Name())
		if err != nil {
			continue
		}
		out = append(out, ResumableRun{
			RunID: s.RunID, UpdatedAt: s.UpdatedAt, Attempts: len(s.Attempts),
			Remaining: s.BudgetUSDRemaining, Prompt: s.OriginalPrompt,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func StartAttempt(state State) int {
	return len(state.Attempts) + 1
}
