// Package webdashboard exposes the production React dashboard build as an
// embed.FS so the `harness` binary can ship a fully self-contained UI.
//
// When `web/dashboard/dist/` only contains the PLACEHOLDER.md (i.e. the
// React build has not run yet), HasIndex() returns false and the HTTP
// server falls back to its built-in HTML page.
package webdashboard

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS returns a filesystem rooted at `dist/` so callers can serve files
// with paths like "index.html" or "assets/foo.js" without the prefix.
func FS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return distFS
	}
	return sub
}

// HasIndex reports whether the embedded build contains an index.html.
// The placeholder-only embed returns false.
func HasIndex() bool {
	_, err := fs.Stat(FS(), "index.html")
	return err == nil
}
