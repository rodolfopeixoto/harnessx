// SPDX-License-Identifier: MIT

package learncmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/execution"
)

// Incremental is the compact rolling summary written to
// .harness/memory/incremental.json after each execution run. It lets
// `harness learn` short-circuit the full replay when nothing meaningful
// changed since the last update.
type Incremental struct {
	GeneratedAt   time.Time      `json:"generated_at"`
	RunsSeen      int            `json:"runs_seen"`
	LastRunID     string         `json:"last_run_id"`
	TokensTotal   int            `json:"tokens_total"`
	CostTotal     float64        `json:"cost_total"`
	ByAdapter     map[string]int `json:"by_adapter"`
	ByStatus      map[string]int `json:"by_status"`
	ByErrorType   map[string]int `json:"by_error_type,omitempty"`
	TokensPerAdpt map[string]int `json:"tokens_per_adapter,omitempty"`
}

func LoadIncremental(root string) (Incremental, error) {
	path := filepath.Join(root, ".harness", "memory", "incremental.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Incremental{
				ByAdapter:     map[string]int{},
				ByStatus:      map[string]int{},
				ByErrorType:   map[string]int{},
				TokensPerAdpt: map[string]int{},
			}, nil
		}
		return Incremental{}, err
	}
	var inc Incremental
	if err := json.Unmarshal(data, &inc); err != nil {
		return Incremental{}, err
	}
	if inc.ByAdapter == nil {
		inc.ByAdapter = map[string]int{}
	}
	if inc.ByStatus == nil {
		inc.ByStatus = map[string]int{}
	}
	if inc.ByErrorType == nil {
		inc.ByErrorType = map[string]int{}
	}
	if inc.TokensPerAdpt == nil {
		inc.TokensPerAdpt = map[string]int{}
	}
	return inc, nil
}

func UpdateIncremental(root string, run execution.Result) (Incremental, string, error) {
	inc, err := LoadIncremental(root)
	if err != nil {
		return Incremental{}, "", err
	}
	if run.RunID == inc.LastRunID {
		return inc, "", nil
	}
	inc.GeneratedAt = time.Now().UTC()
	inc.RunsSeen++
	inc.LastRunID = run.RunID
	inc.TokensTotal += run.InputTokens + run.OutputTokens
	inc.CostTotal += run.EstimatedCostUSD
	inc.ByAdapter[run.AgentID]++
	inc.ByStatus[string(run.Status)]++
	if run.ErrorType != "" {
		inc.ByErrorType[run.ErrorType]++
	}
	inc.TokensPerAdpt[run.AgentID] += run.InputTokens + run.OutputTokens
	path := filepath.Join(root, ".harness", "memory", "incremental.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return inc, "", err
	}
	body, err := json.MarshalIndent(inc, "", "  ")
	if err != nil {
		return inc, "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return inc, "", err
	}
	return inc, path, nil
}

func lookupFixForError(errType string) string {
	switch strings.ToLower(errType) {
	case "budget_exceeded":
		return "Raise --budget-usd OR pin a cheaper adapter (`harness use claude-haiku-4-5`)."
	case "pre_hook_blocked":
		return "Inspect `harness hook list`; the pre-tool-use hook returned non-zero. Fix the script or disable for this run."
	case "worktree_prepare":
		return "Check that the project root is a git repo OR has write permissions for `.harness/worktrees/`."
	}
	return "Inspect the run report (`harness runs report <id>`) and the agent stderr for the root cause."
}
