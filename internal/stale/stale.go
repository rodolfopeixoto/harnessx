// SPDX-License-Identifier: MIT

package stale

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

var TrackedFiles = []string{
	"package.json",
	"package-lock.json",
	"pnpm-lock.yaml",
	"go.mod",
	"go.sum",
	"Dockerfile",
	"docker-compose.yaml",
	"docker-compose.yml",
	"Gemfile",
	"Gemfile.lock",
	"Cargo.toml",
	"Cargo.lock",
	"pyproject.toml",
	"requirements.txt",
}

type Fingerprints struct {
	Files map[string]string `json:"files"`
}

type Entry struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	HashOld string `json:"hash_old,omitempty"`
	HashNew string `json:"hash_new"`
	Reason  string `json:"reason"`
}

const FingerprintFilename = "fingerprints.json"

func fingerprintPath(root string) string {
	return filepath.Join(root, constants.HarnessDir, constants.ProjectSubdir, FingerprintFilename)
}

func Load(root string) (Fingerprints, error) {
	b, err := os.ReadFile(fingerprintPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return Fingerprints{Files: map[string]string{}}, nil
	}
	if err != nil {
		return Fingerprints{}, err
	}
	var fp Fingerprints
	if err := json.Unmarshal(b, &fp); err != nil {
		return Fingerprints{}, err
	}
	if fp.Files == nil {
		fp.Files = map[string]string{}
	}
	return fp, nil
}

func Save(root string, fp Fingerprints) error {
	target := fingerprintPath(root)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(fp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(target, b, 0o644)
}

func Detect(root string) ([]Entry, error) {
	previous, err := Load(root)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, rel := range TrackedFiles {
		full := filepath.Join(root, rel)
		hash, err := hashFile(full)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		old := previous.Files[rel]
		if old == hash {
			continue
		}
		reason := reasonFor(old)
		entries = append(entries, Entry{
			Path:    rel,
			Kind:    kindOf(rel),
			HashOld: old,
			HashNew: hash,
			Reason:  reason,
		})
	}
	return entries, nil
}

func Record(root string) (Fingerprints, error) {
	fp := Fingerprints{Files: map[string]string{}}
	for _, rel := range TrackedFiles {
		hash, err := hashFile(filepath.Join(root, rel))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fp, err
		}
		fp.Files[rel] = hash
	}
	if err := Save(root, fp); err != nil {
		return fp, err
	}
	return fp, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func kindOf(path string) string {
	switch filepath.Base(path) {
	case "Dockerfile", "docker-compose.yaml", "docker-compose.yml":
		return "container"
	case "package.json", "package-lock.json", "pnpm-lock.yaml", "go.mod", "go.sum", "Gemfile", "Gemfile.lock", "Cargo.toml", "Cargo.lock", "pyproject.toml", "requirements.txt":
		return "dependencies"
	default:
		return "config"
	}
}

func reasonFor(old string) string {
	if old == "" {
		return "first time fingerprinted"
	}
	return "content changed since last index"
}
