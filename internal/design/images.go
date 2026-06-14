// SPDX-License-Identifier: MIT

package design

import (
	"encoding/json"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	// register decoders for size extraction
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/ropeixoto/harnessx/internal/platform/hashing"
)

// ImageCache writes per-image analysis JSON under
// .harness/cache/images/<sha256>.json. Phase 7 records hash + format +
// dimensions; a vision-model pass plugs in later by enriching `Detected`.
type ImageCache struct {
	Root string // project root
}

func (c ImageCache) Dir() string { return filepath.Join(c.Root, ".harness", "cache", "images") }

func (c ImageCache) Path(hash string) string {
	return filepath.Join(c.Dir(), hash+".json")
}

// AnalyseAll walks src for image assets, computes a hash, captures
// metadata, and writes one cache JSON per unique hash.
func (c ImageCache) AnalyseAll(src Source) ([]ImageAnalysis, error) {
	if err := os.MkdirAll(c.Dir(), 0o755); err != nil {
		return nil, err
	}
	var out []ImageAnalysis
	_ = filepath.WalkDir(src.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif":
		default:
			return nil
		}
		rel, _ := filepath.Rel(src.Root, p)
		hash, err := hashing.SHA256File(p)
		if err != nil {
			return nil
		}
		analysis := ImageAnalysis{
			Hash: hash, Source: rel, Format: strings.TrimPrefix(ext, "."),
			Label: labelFromPath(rel), CachedAt: time.Now().UTC(),
		}
		info, err := d.Info()
		if err == nil {
			analysis.Bytes = info.Size()
		}
		// best-effort dimensions
		if f, err := os.Open(p); err == nil {
			if cfg, _, err := image.DecodeConfig(f); err == nil {
				analysis.Width, analysis.Height = cfg.Width, cfg.Height
			}
			f.Close()
		}
		// Persist (idempotent — same content => same hash => overwrite).
		if b, err := json.MarshalIndent(analysis, "", "  "); err == nil {
			_ = os.WriteFile(c.Path(hash), b, 0o644)
		}
		out = append(out, analysis)
		return nil
	})
	return out, nil
}

func labelFromPath(rel string) string {
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	return strings.ReplaceAll(base, "_", "-")
}
