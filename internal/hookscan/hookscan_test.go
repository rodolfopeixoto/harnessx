// SPDX-License-Identifier: MIT

package hookscan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeExec(t *testing.T, root, rel, body string, mode os.FileMode) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(body), mode))
}

func TestScan_GitHooks(t *testing.T) {
	root := t.TempDir()
	writeExec(t, root, "scripts/git-hooks/pre-push", "#!/usr/bin/env bash\nexit 0\n", 0o755)
	writeExec(t, root, "scripts/git-hooks/commit-msg", "#!/usr/bin/env bash\nexit 0\n", 0o755)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Len(t, out, 2)
	for _, h := range out {
		require.Equal(t, SourceGit, h.Source)
		require.True(t, h.Blocking)
		require.Equal(t, RiskMedium, h.Risk)
	}
}

func TestScan_ClaudeHook(t *testing.T) {
	root := t.TempDir()
	writeExec(t, root, ".claude/hooks/pre-tool-use.sh", "#!/usr/bin/env bash\n", 0o644)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, SourceClaude, out[0].Source)
	require.Equal(t, StatusDisabled, out[0].Status)
}

func TestScan_EmptyRoot(t *testing.T) {
	_, err := Scan("")
	require.Error(t, err)
}

func TestScan_SkipsNoiseDirs(t *testing.T) {
	root := t.TempDir()
	writeExec(t, root, "node_modules/something/hook.sh", "x", 0o755)
	out, err := Scan(root)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestEventOf_Recognised(t *testing.T) {
	require.Equal(t, "pre-push", eventOf("pre-push"))
	require.Equal(t, "post-merge", eventOf("post-merge"))
	require.Equal(t, "custom", eventOf("anything.sh"))
}
