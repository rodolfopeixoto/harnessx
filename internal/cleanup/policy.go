// SPDX-License-Identifier: MIT

package cleanup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Policy struct {
	Version int           `yaml:"version"`
	Rules   []PolicyRule  `yaml:"rules"`
	Globals PolicyGlobals `yaml:"globals"`
	Source  string        `yaml:"-"`
}

type PolicyRule struct {
	Kind      string   `yaml:"kind"`
	Allowlist []string `yaml:"allowlist"`
	MinAgeH   int      `yaml:"min_age_hours"`
	MaxRisk   string   `yaml:"max_risk"`
}

type PolicyGlobals struct {
	RequireAcknowledgement bool `yaml:"require_acknowledgement"`
}

var ErrPolicyMissing = errors.New("cleanup: no policy match")

func LoadPolicy(root string) (Policy, error) {
	path := filepath.Join(root, constants.HarnessDir, constants.CleanupSubdir, constants.CleanupPolicyFilename)
	return LoadPolicyFile(path)
}

func LoadPolicyFile(path string) (Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return Policy{}, fmt.Errorf("cleanup: parse policy %s: %w", path, err)
	}
	p.Source = path
	if p.Version == 0 {
		p.Version = 1
	}
	return p, nil
}

func (p Policy) Match(f Finding) (PolicyRule, bool) {
	for _, rule := range p.Rules {
		if rule.Kind != "" && rule.Kind != f.Kind {
			continue
		}
		if !pathMatchesAny(rule.Allowlist, f.Path) {
			continue
		}
		if rule.MinAgeH > 0 && ageHours(f) < rule.MinAgeH {
			continue
		}
		if rule.MaxRisk != "" && !riskAllowed(rule.MaxRisk, f.Risk) {
			continue
		}
		return rule, true
	}
	return PolicyRule{}, false
}

func pathMatchesAny(globs []string, path string) bool {
	for _, g := range globs {
		matched, _ := filepath.Match(g, path)
		if matched {
			return true
		}
	}
	return false
}

func ageHours(f Finding) int {
	if f.LastTouched.IsZero() {
		return 0
	}
	delta := nowFn().Sub(f.LastTouched)
	return int(delta.Hours())
}

func riskAllowed(maxRisk string, actual Risk) bool {
	limit := riskOrder(Risk(maxRisk))
	if limit == 0 {
		return false
	}
	return riskOrder(actual) <= limit
}
