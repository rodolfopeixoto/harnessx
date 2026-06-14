// SPDX-License-Identifier: MIT

package http

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	webdashboard "github.com/ropeixoto/harnessx/web/dashboard"
)

// staticOrFallback serves the built React dashboard. Resolution order:
//  1. on-disk dist directory passed via Options.Dist (developer override).
//  2. embedded dist FS shipped with the binary (production single-binary).
//  3. built-in HTML placeholder.
//
// SPA semantics: unknown paths fall back to index.html so client-side
// routing works.
func (s *Server) staticOrFallback(w http.ResponseWriter, r *http.Request) {
	clean := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	if strings.Contains(clean, "..") {
		http.NotFound(w, r)
		return
	}

	if s.dist != "" {
		if r.URL.Path != "/" {
			full := filepath.Join(s.dist, clean)
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				http.ServeFile(w, r, full)
				return
			}
		}
		idx := filepath.Join(s.dist, "index.html")
		if _, err := os.Stat(idx); err == nil {
			http.ServeFile(w, r, idx)
			return
		}
	}

	if webdashboard.HasIndex() {
		dist := webdashboard.FS()
		if r.URL.Path != "/" {
			if b, err := fs.ReadFile(dist, clean); err == nil {
				ext := strings.ToLower(filepath.Ext(clean))
				w.Header().Set("Content-Type", contentTypeFor(ext))
				_, _ = w.Write(b)
				return
			}
		}
		if b, err := fs.ReadFile(dist, "index.html"); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(b)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(builtinHTML))
}

func contentTypeFor(ext string) string {
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	case ".ico":
		return "image/x-icon"
	}
	return "application/octet-stream"
}

const builtinHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>HarnessX Dashboard</title>
<style>
  body { font-family: -apple-system, system-ui, sans-serif; max-width: 920px; margin: 2rem auto; padding: 0 1rem; color: #222; }
  h1 { margin-bottom: 0.25rem; }
  .meta { color: #666; margin-bottom: 1.5rem; }
  table { border-collapse: collapse; width: 100%; font-size: 14px; }
  th, td { border: 1px solid #ddd; padding: 6px 10px; text-align: left; }
  th { background: #f6f6f8; }
  code { background: #f1f1f4; padding: 2px 5px; border-radius: 3px; }
  a { color: #4338CA; }
</style>
</head>
<body>
<h1>HarnessX</h1>
<p class="meta">Dashboard placeholder (run <code>make dashboard-build</code> for the React UI).</p>
<h2>API</h2>
<ul>
  <li><a href="/api/health">/api/health</a></li>
  <li><a href="/api/sessions">/api/sessions</a></li>
  <li><a href="/api/sensors">/api/sensors</a></li>
  <li><a href="/api/agents">/api/agents</a></li>
  <li><a href="/api/cost">/api/cost</a></li>
  <li><a href="/api/logs?tail=20">/api/logs?tail=20</a></li>
  <li><a href="/api/profile">/api/profile</a></li>
  <li><a href="/api/design">/api/design</a></li>
  <li><a href="/api/roadmap">/api/roadmap</a></li>
  <li><a href="/api/features">/api/features</a></li>
  <li><a href="/api/toggles">/api/toggles</a></li>
  <li><a href="/api/memory">/api/memory</a></li>
</ul>
<h2>Recent sessions</h2>
<div id="sessions">loading…</div>
<script>
async function load() {
  const res = await fetch('/api/sessions?limit=20');
  const sessions = await res.json();
  const el = document.getElementById('sessions');
  if (!sessions || !sessions.length) { el.textContent = 'no sessions yet'; return; }
  const rows = sessions.map(s => '<tr><td><code>' + s.ID + '</code></td><td>' + s.Mode + '</td><td>' + s.Status + '</td><td>' + (s.StartedAt || '') + '</td></tr>').join('');
  el.innerHTML = '<table><thead><tr><th>id</th><th>mode</th><th>status</th><th>started</th></tr></thead><tbody>' + rows + '</tbody></table>';
}
load();
</script>
</body>
</html>
`
