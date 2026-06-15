// SPDX-License-Identifier: MIT

package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/version"
)

type Manifest struct {
	Tag             string            `json:"tag,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	HarnessVersion  string            `json:"harness_version"`
	OS              string            `json:"os"`
	Arch            string            `json:"arch"`
	IncludedRoots   []string          `json:"included_roots"`
	IncludedFiles   []FileEntry       `json:"included_files"`
	ExcludedReasons map[string]string `json:"excluded_reasons,omitempty"`
}

type FileEntry struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

const manifestName = ".harness-backup-manifest.json"

func PackName(tag string) string {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	if tag == "" {
		return "harness-backup-" + stamp + ".tar.gz"
	}
	safe := safeTag(tag)
	return "harness-backup-" + stamp + "-" + safe + ".tar.gz"
}

func Pack(projectRoot string, cfg Config, tag, dest string, includeSecrets bool) (Manifest, error) {
	m := Manifest{
		Tag: tag, CreatedAt: time.Now().UTC(),
		HarnessVersion: version.Version,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		IncludedRoots:  cfg.Include,
	}
	f, err := os.Create(dest)
	if err != nil {
		return m, err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer func() { _ = gz.Close() }()
	tw := tar.NewWriter(gz)
	defer func() { _ = tw.Close() }()

	policy := newPolicy(cfg, includeSecrets)
	for _, rel := range cfg.Include {
		root := filepath.Join(projectRoot, rel)
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.Walk(root, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relPath, err := filepath.Rel(projectRoot, p)
			if err != nil {
				return err
			}
			if !policy.Allow(relPath) {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			entry, err := writeTarEntry(tw, p, relPath, info)
			if err != nil {
				return err
			}
			m.IncludedFiles = append(m.IncludedFiles, entry)
			return nil
		})
		if err != nil {
			return m, err
		}
	}
	if err := writeManifest(tw, m); err != nil {
		return m, err
	}
	return m, nil
}

func Unpack(src, dest string, force bool) (Manifest, error) {
	if err := preflightTarget(dest, force); err != nil {
		return Manifest{}, err
	}
	f, err := os.Open(src)
	if err != nil {
		return Manifest{}, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return Manifest{}, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	var manifest Manifest
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return manifest, err
		}
		if m, captured, err := captureManifest(hdr, tr); err != nil {
			return manifest, err
		} else if captured {
			manifest = m
			continue
		}
		if err := writeUnpackedEntry(hdr, tr, dest); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

func preflightTarget(dest string, force bool) error {
	if force {
		return nil
	}
	if entries, err := os.ReadDir(dest); err == nil && len(entries) > 0 {
		return fmt.Errorf("backup: target %s not empty (pass --force)", dest)
	}
	return nil
}

func captureManifest(hdr *tar.Header, tr *tar.Reader) (Manifest, bool, error) {
	if filepath.Clean(hdr.Name) != manifestName {
		return Manifest{}, false, nil
	}
	data, err := io.ReadAll(io.LimitReader(tr, 5*1024*1024))
	if err != nil {
		return Manifest{}, true, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, true, err
	}
	return m, true, nil
}

func writeUnpackedEntry(hdr *tar.Header, tr *tar.Reader, dest string) error {
	clean := filepath.Clean(hdr.Name)
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return fmt.Errorf("backup: refusing path %s", hdr.Name)
	}
	if hdr.Typeflag != tar.TypeReg {
		return nil
	}
	target := filepath.Join(dest, clean)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.CopyN(out, tr, 500*1024*1024); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func writeTarEntry(tw *tar.Writer, abs, rel string, info os.FileInfo) (FileEntry, error) {
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return FileEntry{}, err
	}
	hdr.Name = rel
	if err := tw.WriteHeader(hdr); err != nil {
		return FileEntry{}, err
	}
	f, err := os.Open(abs)
	if err != nil {
		return FileEntry{}, err
	}
	defer f.Close()
	h := sha256.New()
	written, err := io.Copy(io.MultiWriter(tw, h), f)
	if err != nil {
		return FileEntry{}, err
	}
	return FileEntry{Path: rel, Bytes: written, SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func writeManifest(tw *tar.Writer, m Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	hdr := &tar.Header{Name: manifestName, Mode: 0o644, Size: int64(len(data))}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = tw.Write(data)
	return err
}

func safeTag(tag string) string {
	var b strings.Builder
	for _, r := range tag {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := b.String()
	if out == "" {
		return "untagged"
	}
	return out
}
