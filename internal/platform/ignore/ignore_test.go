package ignore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Missing_NoMatches(t *testing.T) {
	m, err := Load(t.TempDir())
	require.NoError(t, err)
	require.False(t, m.Match("any/file.go", false))
}

func TestPatterns(t *testing.T) {
	root := t.TempDir()
	body := `# comment
*.lock
node_modules/
/build
docs/secret.md
`
	require.NoError(t, os.WriteFile(filepath.Join(root, ".harnessignore"), []byte(body), 0o644))
	m, err := Load(root)
	require.NoError(t, err)

	require.True(t, m.Match("yarn.lock", false))
	require.True(t, m.Match("sub/foo.lock", false))
	require.True(t, m.Match("node_modules", true))
	require.False(t, m.Match("node_modules", false), "trailing slash → dir only")
	require.True(t, m.Match("build", false), "anchored to root")
	require.True(t, m.Match("docs/secret.md", false))
	require.False(t, m.Match("other/file.go", false))
}
