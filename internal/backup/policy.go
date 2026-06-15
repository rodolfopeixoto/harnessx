// SPDX-License-Identifier: MIT

package backup

import (
	"path/filepath"
	"strings"
)

type policy struct {
	include        []string
	exclude        []string
	includeSecrets bool
}

func newPolicy(cfg Config, includeSecrets bool) *policy {
	return &policy{include: cfg.Include, exclude: cfg.Exclude, includeSecrets: includeSecrets}
}

var secretPaths = []string{
	".harness/secrets.enc",
	".harness/secret-seed",
}

func isSecretPath(rel string) bool {
	for _, s := range secretPaths {
		if rel == s || strings.HasPrefix(rel, s+"/") {
			return true
		}
	}
	return false
}

func (p *policy) Allow(rel string) bool {
	rel = filepath.ToSlash(rel)
	if isSecretPath(rel) {
		return p.includeSecrets
	}
	for _, ex := range p.exclude {
		if matchRel(ex, rel) {
			return false
		}
	}
	if len(p.include) == 0 {
		return true
	}
	for _, inc := range p.include {
		if matchRel(inc, rel) {
			return true
		}
	}
	return false
}

func matchRel(pattern, target string) bool {
	pattern = filepath.ToSlash(pattern)
	target = filepath.ToSlash(target)
	if pattern == target {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		base := strings.TrimSuffix(pattern, "/**")
		return target == base || strings.HasPrefix(target, base+"/")
	}
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		left := strings.TrimSuffix(parts[0], "/")
		right := strings.TrimPrefix(parts[1], "/")
		return strings.HasPrefix(target, left) && strings.Contains(target, right)
	}
	if strings.HasPrefix(target, pattern+"/") {
		return true
	}
	if ok, _ := filepath.Match(pattern, target); ok {
		return true
	}
	if !strings.Contains(pattern, "/") {
		if ok, _ := filepath.Match(pattern, filepath.Base(target)); ok {
			return true
		}
	}
	return false
}
