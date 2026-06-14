// SPDX-License-Identifier: MIT

package design

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Source is a normalised handle to a design export — either a folder
// already on disk or an extracted temp dir produced from a ZIP.
type Source struct {
	Root    string // directory containing the design files
	Origin  string // user-supplied path (zip or folder)
	Cleanup func() // call when done; no-op for folder sources
}

const maxExtractBytes = 200 * 1024 * 1024 // 200 MiB safety cap for ZIPs

// Resolve accepts a folder path or a `.zip` file and returns a Source
// pointing at a directory containing the design tree. ZIPs are extracted
// into a temp dir; callers must call Source.Cleanup when done.
func Resolve(path string) (Source, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Source{}, fmt.Errorf("design: %w", err)
	}
	if info.IsDir() {
		abs, err := filepath.Abs(path)
		if err != nil {
			return Source{}, err
		}
		return Source{Root: abs, Origin: path, Cleanup: func() {}}, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ".zip") {
		return Source{}, fmt.Errorf("design: unsupported source %q (want folder or .zip)", path)
	}
	dir, err := extractZip(path)
	if err != nil {
		return Source{}, err
	}
	return Source{
		Root: dir, Origin: path,
		Cleanup: func() { _ = os.RemoveAll(dir) },
	}, nil
}

// extractZip safely unpacks zipPath into a temp dir. Rejects entries
// whose target paths would escape the destination (zip-slip) or whose
// uncompressed sizes blow the cap.
func extractZip(zipPath string) (string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("design: open zip: %w", err)
	}
	defer zr.Close()

	dest, err := os.MkdirTemp("", "harnessx-design-*")
	if err != nil {
		return "", err
	}
	var total int64
	for _, f := range zr.File {
		// Reject absolute paths and `..` segments before joining.
		clean := filepath.Clean(f.Name)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") || strings.Contains(clean, ".."+string(os.PathSeparator)) {
			_ = os.RemoveAll(dest)
			return "", fmt.Errorf("design: unsafe entry %q", f.Name)
		}
		target := filepath.Join(dest, clean)
		// Guard against join-resolved escape (Windows compat etc.).
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != dest {
			_ = os.RemoveAll(dest)
			return "", fmt.Errorf("design: entry escapes destination: %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				_ = os.RemoveAll(dest)
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = os.RemoveAll(dest)
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			_ = os.RemoveAll(dest)
			return "", err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			_ = os.RemoveAll(dest)
			return "", err
		}
		n, err := io.CopyN(out, rc, maxExtractBytes-total+1)
		if err != nil && !errors.Is(err, io.EOF) {
			rc.Close()
			out.Close()
			_ = os.RemoveAll(dest)
			return "", fmt.Errorf("design: copy %s: %w", f.Name, err)
		}
		total += n
		rc.Close()
		out.Close()
		if total > maxExtractBytes {
			_ = os.RemoveAll(dest)
			return "", fmt.Errorf("design: extracted size exceeds %d bytes", maxExtractBytes)
		}
	}
	return dest, nil
}
