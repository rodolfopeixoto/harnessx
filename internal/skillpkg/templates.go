// SPDX-License-Identifier: MIT

// Package skillpkg bundles short, deterministic skill snippets harness
// prefixes onto agent prompts when the router selects them. Each
// template is a markdown body keyed by `mode` and short slug.
package skillpkg

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed templates/*.md
var bundled embed.FS

type Template struct {
	Name        string
	Mode        string
	Description string
	Body        string
}

func List() ([]Template, error) {
	var out []Template
	err := fs.WalkDir(bundled, "templates", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		t, err := loadByPath(path)
		if err != nil {
			return err
		}
		out = append(out, t)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func Load(name string) (Template, error) {
	return loadByPath("templates/" + name + ".md")
}

func loadByPath(path string) (Template, error) {
	data, err := bundled.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("skillpkg: not bundled: %s", path)
	}
	base := strings.TrimSuffix(strings.TrimPrefix(path, "templates/"), ".md")
	t := Template{Name: base, Body: string(data)}
	t.Mode = parseHeader(data, "<!-- mode:")
	t.Description = parseHeader(data, "<!-- description:")
	return t, nil
}

func parseHeader(body []byte, prefix string) string {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			rest = strings.TrimSuffix(rest, "-->")
			return strings.TrimSpace(rest)
		}
	}
	return ""
}
