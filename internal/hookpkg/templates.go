// SPDX-License-Identifier: MIT

// Package hookpkg bundles ready-to-drop hook scripts harness can
// install with one command. Each template is a self-contained bash
// script that respects the HARNESS_RUN_ID + HARNESS_AGENT env vars
// the Executor injects.
package hookpkg

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed templates/*.sh
var bundled embed.FS

type Template struct {
	Name        string
	Event       string
	Description string
	Body        []byte
}

func List() ([]Template, error) {
	var out []Template
	err := fs.WalkDir(bundled, "templates", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if !strings.HasSuffix(path, ".sh") {
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
	return loadByPath("templates/" + name + ".sh")
}

func loadByPath(path string) (Template, error) {
	data, err := bundled.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("hookpkg: not bundled: %s", path)
	}
	base := strings.TrimSuffix(strings.TrimPrefix(path, "templates/"), ".sh")
	t := Template{Name: base, Body: data}
	t.Event = inferEvent(base, data)
	t.Description = inferDescription(data)
	return t, nil
}

func inferEvent(name string, body []byte) string {
	if e := parseHeader(body, "# event:"); e != "" {
		return e
	}
	switch {
	case strings.HasPrefix(name, "pre-tool-use"):
		return "pre-tool-use"
	case strings.HasPrefix(name, "post-tool-use"):
		return "post-tool-use"
	case strings.HasPrefix(name, "session-start"):
		return "session-start"
	case strings.HasPrefix(name, "session-end"):
		return "session-end"
	}
	return ""
}

func inferDescription(body []byte) string {
	return parseHeader(body, "# description:")
}

func parseHeader(body []byte, prefix string) string {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}
