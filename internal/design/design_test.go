package design

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
}

func newFolderSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "index.html", `<!doctype html><html><head><title>Home</title></head><body>
<a href="/signup">Sign up</a>
<button class="UiButton" onclick="go()">Go</button>
</body></html>`)
	writeFile(t, root, "signup.html", `<!doctype html><html><head><title>Sign up</title></head><body>
<form onsubmit="signup()">
  <input class="ui-input"/>
  <div class="loading">Loading…</div>
  <button class="UiButton">Submit</button>
</form>
</body></html>`)
	writeFile(t, root, "styles/site.css", `body { background: #ffffff; color: #7C3AED; padding: 16px; }
@media (max-width: 640px) { body { padding: 8px; } }`)
	writeFile(t, root, "components/Button.tsx", `export const Button = () => null;`)
	writeFile(t, root, "assets/logo.svg", `<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	return root
}

func TestInventory_DetectsPagesComponentsAssetsAndStyles(t *testing.T) {
	root := newFolderSource(t)
	src, err := Resolve(root)
	require.NoError(t, err)
	defer src.Cleanup()

	m, err := Inventory(src)
	require.NoError(t, err)
	require.Len(t, m.Pages, 2)
	require.NotEmpty(t, m.Components)
	require.Contains(t, m.Assets, "assets/logo.svg")
	require.Contains(t, m.Styles.Colors, "#ffffff")
	require.Contains(t, m.Styles.Spacing, "16px")
	require.NotEmpty(t, m.Responsive)
}

func TestFeatureMap_SignupRequiresBackend(t *testing.T) {
	root := newFolderSource(t)
	src, _ := Resolve(root)
	defer src.Cleanup()
	m, _ := Inventory(src)
	fm := BuildFeatureMap(m)
	var found bool
	for id, f := range fm.Features {
		if id == "feature.signup" {
			require.True(t, f.BackendRequired)
			require.Equal(t, StatusMock, f.Status)
			require.NotEmpty(t, f.APIContract)
			found = true
		}
	}
	require.True(t, found, "expected feature.signup in %v", fm.Features)
}

func TestRoadmap_HasFivePhases(t *testing.T) {
	root := newFolderSource(t)
	src, _ := Resolve(root)
	defer src.Cleanup()
	m, _ := Inventory(src)
	fm := BuildFeatureMap(m)
	r := BuildRoadmap(fm)
	require.Len(t, r.Phases, 5)
	require.Equal(t, "MVP 0", r.Phases[0].Name)
}

func TestExtractZip_RejectsPathTraversal(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "bad.zip")
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	hdr, err := w.Create("../escape.txt")
	require.NoError(t, err)
	_, _ = hdr.Write([]byte("oops"))
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	_, err = Resolve(zipPath)
	require.Error(t, err)
}

func TestExtractZip_HappyPath(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "ok.zip")
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	wr, _ := w.Create("index.html")
	_, _ = wr.Write([]byte("<html><title>x</title></html>"))
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	src, err := Resolve(zipPath)
	require.NoError(t, err)
	defer src.Cleanup()
	require.DirExists(t, src.Root)
}

// guard against accidental imports
var _ = bytes.Buffer{}
