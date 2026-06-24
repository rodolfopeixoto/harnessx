// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleRegistry = `# Servers

## Reference servers

- **[Filesystem](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem)** - Local filesystem access.
- **[GitHub](https://github.com/modelcontextprotocol/servers/tree/main/src/github)** - GitHub repo browsing.
- **[Slack](https://github.com/modelcontextprotocol/servers/tree/main/src/slack)** - Slack channel reading.

## Third party

* **[Brave Search](https://github.com/modelcontextprotocol/servers/tree/main/src/brave-search)** - Brave Search results.
`

func TestSearchRegistryFiltersBySubstring(t *testing.T) {
	hits := searchRegistry([]byte(sampleRegistry), "slack")
	if len(hits) != 1 || hits[0].Name != "Slack" {
		t.Errorf("want one Slack hit, got %+v", hits)
	}
}

func TestSearchRegistryEmptyTermReturnsAll(t *testing.T) {
	hits := searchRegistry([]byte(sampleRegistry), "")
	if len(hits) != 4 {
		t.Errorf("want 4 hits, got %d (%+v)", len(hits), hits)
	}
}

func TestSearchRegistrySortedByName(t *testing.T) {
	hits := searchRegistry([]byte(sampleRegistry), "")
	for i := 1; i < len(hits); i++ {
		if hits[i-1].Name > hits[i].Name {
			t.Errorf("not sorted: %s > %s", hits[i-1].Name, hits[i].Name)
		}
	}
}

func TestLoadRegistryUsesFreshCache(t *testing.T) {
	root := t.TempDir()
	cachePath := filepath.Join(root, ".harness", "cache", "mcp-registry.md")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, []byte("cached body"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := fetchRegistry
	t.Cleanup(func() { fetchRegistry = orig })
	fetchRegistry = func(string) ([]byte, error) { t.Fatal("fetch must not be called when cache fresh"); return nil, nil }
	body, src, err := loadRegistry(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "cached body" {
		t.Errorf("cached body wrong: %s", body)
	}
	if !strings.Contains(src, "cached") {
		t.Errorf("source label wrong: %s", src)
	}
}

func TestLoadRegistryRefreshForcesFetch(t *testing.T) {
	root := t.TempDir()
	cachePath := filepath.Join(root, ".harness", "cache", "mcp-registry.md")
	_ = os.MkdirAll(filepath.Dir(cachePath), 0o755)
	_ = os.WriteFile(cachePath, []byte("stale"), 0o644)
	orig := fetchRegistry
	t.Cleanup(func() { fetchRegistry = orig })
	fetchRegistry = func(string) ([]byte, error) { return []byte("fresh"), nil }
	body, src, err := loadRegistry(root, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "fresh" {
		t.Errorf("want fresh, got %s", body)
	}
	if !strings.Contains(src, "fresh") {
		t.Errorf("label wrong: %s", src)
	}
}

func TestMCPSearchCmdRunsAgainstStubbedRegistry(t *testing.T) {
	root := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	orig := fetchRegistry
	t.Cleanup(func() { fetchRegistry = orig })
	fetchRegistry = func(string) ([]byte, error) { return []byte(sampleRegistry), nil }

	c := mcpSearchCmd()
	var buf bytes.Buffer
	c.SetOut(&buf)
	c.SetErr(&buf)
	c.SetArgs([]string{"github"})
	if err := c.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "GitHub") {
		t.Errorf("registry hit missing: %s", got)
	}
	if !strings.Contains(got, "registry source:") {
		t.Errorf("source line missing: %s", got)
	}
}
