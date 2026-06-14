// SPDX-License-Identifier: MIT

package auditrun

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCLIFlows_CapturesExitCodeAndOutput(t *testing.T) {
	flows := []CLIFlow{
		{Name: "echo ok", Args: []string{"hello"}},
		{Name: "false", Args: []string{}},
	}
	binEcho := writeStubBin(t, "echo-bin", "#!/usr/bin/env bash\necho \"$1\"\nexit 0\n")
	binFail := writeStubBin(t, "fail-bin", "#!/usr/bin/env bash\necho boom >&2\nexit 7\n")
	report := CLIReport{}
	report.Flows = append(report.Flows, runFlow(context.Background(), binEcho, "", flows[0]))
	report.Flows = append(report.Flows, runFlow(context.Background(), binFail, "", flows[1]))
	require.Equal(t, 0, report.Flows[0].ExitCode)
	require.Contains(t, report.Flows[0].Stdout, "hello")
	require.Equal(t, 7, report.Flows[1].ExitCode)
	require.Contains(t, report.Flows[1].Stderr, "boom")
}

func TestPrepareCLITmpRoot_InitialisesGit(t *testing.T) {
	dir, cleanup, err := PrepareCLITmpRoot()
	require.NoError(t, err)
	defer cleanup()
	_, err = os.Stat(filepath.Join(dir, ".git"))
	require.NoError(t, err)
}

func TestTruncateOutput_RespectsLimit(t *testing.T) {
	large := make([]byte, 8*1024)
	for i := range large {
		large[i] = 'a'
	}
	out := truncateOutput(string(large))
	require.LessOrEqual(t, len(out), 5*1024)
	require.Contains(t, out, "truncated")
}

func TestDefaultCLIFlows_NotEmpty(t *testing.T) {
	require.NotEmpty(t, DefaultCLIFlows())
}

func writeStubBin(t *testing.T, name, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o755))
	return path
}
