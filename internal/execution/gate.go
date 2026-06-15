// SPDX-License-Identifier: MIT

package execution

import (
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/autonomy"
)

// ClassifyRisk returns "high" when any changed path matches a sensitive
// class (deps, Dockerfile, migrations, secrets, CI config, security
// config, env files), "low" otherwise.
func ClassifyRisk(changed []string) string {
	for _, p := range changed {
		if isHighRiskPath(p) {
			return "high"
		}
	}
	return "low"
}

func isHighRiskPath(p string) bool {
	low := strings.ToLower(filepath.ToSlash(p))
	base := strings.ToLower(filepath.Base(p))
	highBases := []string{
		"dockerfile", "docker-compose.yml", "docker-compose.yaml",
		"go.mod", "go.sum", "package.json", "package-lock.json",
		"pnpm-lock.yaml", "yarn.lock", "cargo.toml", "cargo.lock",
		"requirements.txt", "poetry.lock", "pyproject.toml", "gemfile.lock",
		".env", ".env.local", ".env.production",
	}
	for _, b := range highBases {
		if base == b {
			return true
		}
	}
	highSubs := []string{
		"/migrations/", "/.github/workflows/", "/secrets/",
		".harness/config/autonomy.yaml",
		".harness/config/routes.yaml",
	}
	for _, s := range highSubs {
		if strings.Contains(low, s) {
			return true
		}
	}
	return false
}

// GateApply decides whether the run may write its diff back to the
// project root. Inputs: autonomy level, risk class (from ClassifyRisk),
// sensor outcomes. Output: allow/require_approval/deny + reason.
func GateApply(level AutonomyLevel, risk string, sensors []SensorOutcome) (autonomy.Decision, string) {
	for _, s := range sensors {
		if s.Status == "failed" {
			return autonomy.DecisionDeny, "blocking sensor failed: " + s.ID
		}
	}
	op := autonomy.OpExecuteLowRisk
	if risk == "high" {
		op = autonomy.OpExecuteHighRisk
	}
	dec, err := autonomy.Gate(autonomy.Level(level), op)
	if err != nil {
		return autonomy.DecisionDeny, err.Error()
	}
	return dec, ""
}

// GateApplyWithPolicy extends GateApply with per-project path globs.
// If any changed path matches a deny_paths glob the decision becomes
// Deny + the offending paths in the reason. If allow_paths is non-empty
// and any changed path does not match, the decision is downgraded to
// require_approval. Otherwise the autonomy.Level decision stands.
func GateApplyWithPolicy(level AutonomyLevel, risk string, sensors []SensorOutcome, policy autonomy.Policy, changed []string) (autonomy.Decision, string) {
	base, reason := GateApply(level, risk, sensors)
	if base == autonomy.DecisionDeny {
		return base, reason
	}
	var denied []string
	var unmatched []string
	for _, p := range changed {
		switch policy.MatchPath(p) {
		case autonomy.DecisionDeny:
			denied = append(denied, p)
		case autonomy.DecisionApproval:
			unmatched = append(unmatched, p)
		}
	}
	if len(denied) > 0 {
		return autonomy.DecisionDeny, "policy deny: " + joinShort(denied)
	}
	if len(unmatched) > 0 && len(policy.AllowPaths) > 0 {
		return autonomy.DecisionApproval, "policy approval (not in allow_paths): " + joinShort(unmatched)
	}
	return base, reason
}

func joinShort(paths []string) string {
	if len(paths) > 3 {
		return paths[0] + ", " + paths[1] + ", " + paths[2] + ", ..."
	}
	return joinStrings(paths, ", ")
}

func joinStrings(items []string, sep string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	out := items[0]
	for _, s := range items[1:] {
		out += sep + s
	}
	return out
}
