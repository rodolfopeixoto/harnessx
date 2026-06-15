// SPDX-License-Identifier: MIT

package autonomy

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Policy is the per-project autonomy config loaded from
// .harness/config/autonomy.yaml. It extends the global Level with
// optional path / command allow / deny globs. Empty allow_paths means
// "anything not denied is allowed"; non-empty allow_paths means a path
// must match one of them to be considered allow.
type Policy struct {
	Level          string   `yaml:"level"`
	AllowPaths     []string `yaml:"allow_paths,omitempty"`
	DenyPaths      []string `yaml:"deny_paths,omitempty"`
	AllowCommands  []string `yaml:"allow_commands,omitempty"`
	DenyCommands   []string `yaml:"deny_commands,omitempty"`
	HookBypassFull bool     `yaml:"full_loop_bypasses_hooks"`
}

const policyRelPath = ".harness/config/autonomy.yaml"

// LoadPolicy reads the policy file under projectRoot. Missing file is
// not an error; callers receive the zero-value Policy.
func LoadPolicy(projectRoot string) (Policy, error) {
	path := filepath.Join(projectRoot, policyRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Policy{}, nil
		}
		return Policy{}, err
	}
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Policy{}, err
	}
	return p, nil
}

// MatchPath returns Allow when no deny_paths match and (allow_paths
// empty or one matches). Returns Deny on first deny_paths match. The
// pattern syntax is Go filepath.Match plus a leading "**/" prefix that
// matches across directories.
func (p Policy) MatchPath(rel string) Decision {
	rel = filepath.ToSlash(rel)
	for _, deny := range p.DenyPaths {
		if matchGlob(deny, rel) {
			return DecisionDeny
		}
	}
	if len(p.AllowPaths) == 0 {
		return DecisionAllow
	}
	for _, allow := range p.AllowPaths {
		if matchGlob(allow, rel) {
			return DecisionAllow
		}
	}
	return DecisionApproval
}

// MatchCommand returns Allow when no deny_commands prefix matches and
// (allow_commands empty or one matches as prefix). Deny on first
// deny_commands prefix match. Matching is prefix-based on a normalised
// command string ("git push --force" matches "git push --force --tags").
func (p Policy) MatchCommand(cmd string) Decision {
	c := strings.TrimSpace(cmd)
	for _, deny := range p.DenyCommands {
		if strings.HasPrefix(c, deny) {
			return DecisionDeny
		}
	}
	if len(p.AllowCommands) == 0 {
		return DecisionAllow
	}
	for _, allow := range p.AllowCommands {
		if strings.HasPrefix(c, allow) {
			return DecisionAllow
		}
	}
	return DecisionApproval
}

// matchGlob supports `**/foo`, `foo/**`, `foo/**/bar`, and plain
// filepath.Match patterns.
func matchGlob(pattern, target string) bool {
	pattern = filepath.ToSlash(pattern)
	if pattern == target {
		return true
	}
	if strings.Contains(pattern, "**") {
		return matchDoubleStar(pattern, target)
	}
	ok, _ := filepath.Match(pattern, target)
	if ok {
		return true
	}
	// also match basename so "*.env" works against "config/.env"
	if !strings.Contains(pattern, "/") {
		ok2, _ := filepath.Match(pattern, filepath.Base(target))
		return ok2
	}
	return false
}

func matchDoubleStar(pattern, target string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return false
	}
	left := strings.TrimSuffix(parts[0], "/")
	right := strings.TrimPrefix(parts[1], "/")
	switch {
	case left == "" && right == "":
		return true
	case left == "":
		return strings.Contains(target, right) || strings.HasSuffix(target, right)
	case right == "":
		return strings.HasPrefix(target, left+"/") || target == left
	default:
		return strings.HasPrefix(target, left+"/") && strings.Contains(target, right)
	}
}
