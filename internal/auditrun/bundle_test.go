// SPDX-License-Identifier: MIT

package auditrun

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func TestWriteBundle_ContainsIndexAndArtifacts(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "json"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(base, "report"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "json", "summary.json"), []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "report", "audit.html"), []byte(`<html></html>`), 0o644))

	target := filepath.Join(base, "report", constants.AuditBundleFile)
	size, err := WriteBundle(base, target)
	require.NoError(t, err)
	require.Greater(t, size, int64(0))

	zr, err := zip.OpenReader(target)
	require.NoError(t, err)
	defer zr.Close()
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	require.True(t, names[constants.AuditBundleIndex], "BUNDLE_INDEX.md missing")
	require.True(t, names["json/summary.json"], "summary.json missing")
	require.True(t, names["report/audit.html"], "audit.html missing")
	require.False(t, names["report/"+constants.AuditBundleFile], "bundle should not include itself")
}

func TestBuildInventory_CountsByExtension(t *testing.T) {
	root := t.TempDir()
	must := func(p, body string) {
		full := filepath.Join(root, p)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
	}
	must("cmd/main.go", "package main\n")
	must("cmd/main_test.go", "package main\n")
	must("web/dashboard/src/App.tsx", "export {}\n")
	must("web/dashboard/src/App.test.tsx", "export {}\n")
	must("scripts/test-foo.sh", "#!/usr/bin/env bash\n")
	must("scripts/e2e-phase1.sh", "#!/usr/bin/env bash\n")
	must(".harness/artifacts/specs/p01.md", "# spec\n")

	inv := BuildInventory(root)
	require.Equal(t, 2, inv.GoFiles)
	require.Equal(t, 1, inv.GoTestFiles)
	require.Equal(t, 2, inv.TSXFiles)
	require.Equal(t, 1, inv.TSXTestFiles)
	require.Equal(t, 2, inv.ShellScripts)
	require.Equal(t, 1, inv.ShellTests)
	require.Equal(t, 1, inv.E2EScripts)
}

func TestCountLines_Empty(t *testing.T) {
	require.Equal(t, 0, countLines(filepath.Join(t.TempDir(), "missing")))
	tmp := filepath.Join(t.TempDir(), "f")
	require.NoError(t, os.WriteFile(tmp, []byte("a\nb\nc\n"), 0o644))
	require.Equal(t, 4, countLines(tmp))
}
