// SPDX-License-Identifier: MIT

package mcpscan

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Transport string

const (
	TransportStdio   Transport = "stdio"
	TransportHTTP    Transport = "http"
	TransportUnknown Transport = "unknown"
)

type McpServer struct {
	Name       string    `json:"name"`
	ConfigPath string    `json:"config_path"`
	Source     string    `json:"source"`
	Transport  Transport `json:"transport"`
	Command    string    `json:"command,omitempty"`
	URL        string    `json:"url,omitempty"`
	Risk       string    `json:"risk"`
}

const (
	SourceHarness = "harness"
	SourceClaude  = "claude"
	SourceCodex   = "codex"
	SourceGemini  = "gemini"
	SourceKimi    = "kimi"
	SourceProject = "project"
)

const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

var scanPrefixes = []string{
	".harness/mcp",
	".claude",
	".codex",
	".gemini",
	".kimi",
}

func Scan(root string) ([]McpServer, error) {
	if root == "" {
		return nil, errors.New("mcpscan: empty root")
	}
	seen := map[string]struct{}{}
	var out []McpServer
	for _, prefix := range scanPrefixes {
		base := filepath.Join(root, prefix)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !isConfigExt(path) {
				return nil
			}
			if _, dup := seen[path]; dup {
				return nil
			}
			seen[path] = struct{}{}
			out = append(out, parseFile(root, path)...)
			return nil
		})
	}
	if err := walkForMCPNamed(root, seen, &out); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return out, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func isConfigExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yml", ".yaml":
		return true
	}
	return false
}

func walkForMCPNamed(root string, seen map[string]struct{}, out *[]McpServer) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "dist" || base == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		base := filepath.Base(path)
		if !strings.Contains(strings.ToLower(base), "mcp") {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".json" && ext != ".yml" && ext != ".yaml" {
			return nil
		}
		if _, dup := seen[path]; dup {
			return nil
		}
		seen[path] = struct{}{}
		*out = append(*out, parseFile(root, path)...)
		return nil
	})
}

func parseFile(root, path string) []McpServer {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	source := sourceOf(root, path)
	var generic map[string]any
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(body, &generic); err != nil {
			return nil
		}
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(body, &generic); err != nil {
			return nil
		}
	default:
		return nil
	}
	servers := extractServers(generic)
	var out []McpServer
	for name, entry := range servers {
		out = append(out, McpServer{
			Name:       name,
			ConfigPath: path,
			Source:     source,
			Transport:  transportFor(entry),
			Command:    stringField(entry, "command"),
			URL:        stringField(entry, "url"),
			Risk:       riskFor(entry),
		})
	}
	return out
}

func extractServers(generic map[string]any) map[string]map[string]any {
	out := map[string]map[string]any{}
	if generic == nil {
		return out
	}
	candidates := []string{"mcpServers", "mcp_servers", "servers", "mcps"}
	for _, key := range candidates {
		if raw, ok := generic[key]; ok {
			if m, ok := raw.(map[string]any); ok {
				for name, entry := range m {
					if entryMap, ok := entry.(map[string]any); ok {
						out[name] = entryMap
					}
				}
			}
		}
	}
	if len(out) == 0 && stringField(generic, "command") != "" {
		out["root"] = generic
	}
	return out
}

func transportFor(entry map[string]any) Transport {
	switch strings.ToLower(stringField(entry, "transport")) {
	case "stdio":
		return TransportStdio
	case "http", "https", "sse":
		return TransportHTTP
	}
	if stringField(entry, "command") != "" {
		return TransportStdio
	}
	if stringField(entry, "url") != "" {
		return TransportHTTP
	}
	return TransportUnknown
}

func riskFor(entry map[string]any) string {
	if env, ok := entry["env"].(map[string]any); ok && len(env) > 0 {
		return RiskMedium
	}
	if stringField(entry, "url") != "" {
		return RiskMedium
	}
	return RiskLow
}

func stringField(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func sourceOf(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return SourceProject
	}
	switch {
	case strings.HasPrefix(rel, ".harness/"):
		return SourceHarness
	case strings.HasPrefix(rel, ".claude/"):
		return SourceClaude
	case strings.HasPrefix(rel, ".codex/"):
		return SourceCodex
	case strings.HasPrefix(rel, ".gemini/"):
		return SourceGemini
	case strings.HasPrefix(rel, ".kimi/"):
		return SourceKimi
	}
	return SourceProject
}
