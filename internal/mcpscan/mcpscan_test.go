// SPDX-License-Identifier: MIT

package mcpscan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
}

func TestScan_HarnessJSON(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".harness/mcp/project.json", `{"mcpServers":{"filesystem":{"command":"npx -y mcp-fs","transport":"stdio"},"github":{"url":"https://mcp.github.com","env":{"TOKEN":"x"}}}}`)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Len(t, out, 2)
	names := map[string]McpServer{}
	for _, s := range out {
		names[s.Name] = s
	}
	require.Equal(t, TransportStdio, names["filesystem"].Transport)
	require.Equal(t, TransportHTTP, names["github"].Transport)
	require.Equal(t, RiskMedium, names["github"].Risk)
	require.Equal(t, SourceHarness, names["filesystem"].Source)
}

func TestScan_ClaudeAndCodex(t *testing.T) {
	root := t.TempDir()
	write(t, root, ".claude/mcp.json", `{"mcpServers":{"fs":{"command":"npx mcp-fs"}}}`)
	write(t, root, ".codex/settings/mcp-config.json", `{"servers":{"weather":{"url":"https://wx.example/mcp"}}}`)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Len(t, out, 2)
	bySource := map[string]McpServer{}
	for _, s := range out {
		bySource[s.Source] = s
	}
	require.Equal(t, "fs", bySource[SourceClaude].Name)
	require.Equal(t, "weather", bySource[SourceCodex].Name)
}

func TestScan_GenericMcpFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, "configs/mcp-team.yaml", "servers:\n  search:\n    command: \"npx search\"\n    env:\n      KEY: x\n")
	out, err := Scan(root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "search", out[0].Name)
	require.Equal(t, RiskMedium, out[0].Risk)
}

func TestScan_EmptyRoot(t *testing.T) {
	_, err := Scan("")
	require.Error(t, err)
}

func TestScan_SkipsNoiseDirs(t *testing.T) {
	root := t.TempDir()
	write(t, root, "node_modules/mcp/x.json", `{"mcpServers":{"junk":{"command":"x"}}}`)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestRiskFor_DefaultsLow(t *testing.T) {
	require.Equal(t, RiskLow, riskFor(map[string]any{"command": "x"}))
	require.Equal(t, RiskMedium, riskFor(map[string]any{"command": "x", "env": map[string]any{"K": "v"}}))
	require.Equal(t, RiskMedium, riskFor(map[string]any{"url": "https://x"}))
}

func TestTransportFor_HandlesAliases(t *testing.T) {
	require.Equal(t, TransportStdio, transportFor(map[string]any{"transport": "stdio"}))
	require.Equal(t, TransportHTTP, transportFor(map[string]any{"transport": "sse"}))
	require.Equal(t, TransportUnknown, transportFor(map[string]any{}))
}
