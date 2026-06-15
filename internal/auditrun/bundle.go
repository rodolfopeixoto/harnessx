// SPDX-License-Identifier: MIT

package auditrun

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func WriteBundle(base, target string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, err
	}
	out, err := os.Create(target)
	if err != nil {
		return 0, err
	}
	zw := zip.NewWriter(out)
	if err := writeIndex(zw, base); err != nil {
		_ = zw.Close()
		_ = out.Close()
		return 0, err
	}
	if err := filepath.WalkDir(base, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, constants.AuditBundleFile) {
			return nil
		}
		rel, _ := filepath.Rel(base, path)
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		fh, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fh.Close()
		_, err = io.Copy(w, fh)
		return err
	}); err != nil {
		_ = zw.Close()
		_ = out.Close()
		return 0, err
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		return 0, err
	}
	if err := out.Close(); err != nil {
		return 0, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func writeIndex(zw *zip.Writer, base string) error {
	w, err := zw.Create(constants.AuditBundleIndex)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# audit bundle\n\n")
	b.WriteString("Layout:\n\n")
	b.WriteString("- `json/feature-map.json` — features asserted this run.\n")
	b.WriteString("- `json/results.json` — per (feature × viewport) outcome.\n")
	b.WriteString("- `json/summary.json` — aggregated counts + pass rate.\n")
	b.WriteString("- `json/cli-flows.json` — every CLI subcommand run + exit code + stdout/stderr.\n")
	b.WriteString("- `json/inventory.json` — code/test/spec file counts + largest files.\n")
	b.WriteString("- `json/visual-diff.json`, `json/layout-metrics.json`, `json/network-errors.json`, `json/console-errors.json`, `json/missing-selectors.json` — per-axis diagnostic detail.\n")
	b.WriteString("- `current/screenshots/` — actual screenshots captured this run.\n")
	b.WriteString("- `reference/screenshots/` — design handoff reference screenshots when AUDIT_VISUAL=1.\n")
	b.WriteString("- `diff/` — pixel diff PNGs when reference is available.\n")
	b.WriteString("- `report/audit.html` — human-readable report.\n")
	b.WriteString("- `report/audit.pdf` — printable copy when Playwright PDF emitter ran.\n")
	b.WriteString("- `report/fix-backlog.md` — ranked P0/P1/P2/P3 backlog with reproduction commands.\n")
	b.WriteString("- `run.log` — RFC3339 timestamped runner log.\n\n")
	b.WriteString("Source tree: ")
	b.WriteString(base)
	b.WriteString("\n")
	_, err = io.WriteString(w, b.String())
	if err != nil {
		return fmt.Errorf("audit bundle index: %w", err)
	}
	return nil
}
