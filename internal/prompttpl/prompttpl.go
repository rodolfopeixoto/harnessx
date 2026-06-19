// SPDX-License-Identifier: MIT

// Package prompttpl stores reusable chat prompt templates under
// .harness/prompts/<name>.md so users can replay them via /prompt.
package prompttpl

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,40}$`)

func dir(root string) string {
	return filepath.Join(root, ".harness", "prompts")
}

func path(root, name string) string {
	return filepath.Join(dir(root), name+".md")
}

// ValidName reports whether name is safe to interpolate into a file
// path. Lowercase alphanumeric plus underscore/dash, ≤40 chars, must
// start with a letter or digit.
func ValidName(name string) bool { return nameRe.MatchString(name) }

func Save(root, name, body string) error {
	if !ValidName(name) {
		return fmt.Errorf("prompttpl: invalid name %q (lowercase alnum, _ or -, ≤40 chars)", name)
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("prompttpl: empty body")
	}
	if err := os.MkdirAll(dir(root), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path(root, name), []byte(body), 0o644)
}

func Load(root, name string) (string, error) {
	body, err := os.ReadFile(path(root, name))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func List(root string) ([]string, error) {
	entries, err := os.ReadDir(dir(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		out = append(out, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(out)
	return out, nil
}
