// SPDX-License-Identifier: MIT

package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

// ApplyResult summarises the side-effects of a successful Apply.
type ApplyResult struct {
	Written []string
	Deleted []string
}

// Apply commits the ops atomically. Plan files are staged in a temp dir +
// renamed on success; rollback is a single RemoveAll. Path traversal outside
// root is rejected.
func Apply(_ context.Context, root string, ops []domain.FileOp) (ApplyResult, error) {
	if len(ops) == 0 {
		return ApplyResult{}, nil
	}
	stage, err := os.MkdirTemp("", "harnessx-catalog-")
	if err != nil {
		return ApplyResult{}, fmt.Errorf("catalog: stage: %w", err)
	}
	defer os.RemoveAll(stage)

	var res ApplyResult
	for _, op := range ops {
		clean, err := safeJoin(root, op.Path)
		if err != nil {
			return res, err
		}
		if err := applyOne(stage, clean, op, &res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func applyOne(stage, clean string, op domain.FileOp, res *ApplyResult) error {
	switch op.Op {
	case domain.FileMkdir:
		if err := os.MkdirAll(clean, 0o755); err != nil {
			return fmt.Errorf("catalog: mkdir %s: %w", clean, err)
		}
		res.Written = append(res.Written, clean)
	case domain.FileCreate:
		if _, err := os.Stat(clean); err == nil {
			return fmt.Errorf("catalog: %s already exists (use overwrite or remove first)", clean)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("catalog: stat %s: %w", clean, err)
		}
		if err := writeStaged(stage, clean, op.Body); err != nil {
			return err
		}
		res.Written = append(res.Written, clean)
	case domain.FileOverwrite:
		if err := writeStaged(stage, clean, op.Body); err != nil {
			return err
		}
		res.Written = append(res.Written, clean)
	case domain.FileAppend:
		existing, _ := os.ReadFile(clean)
		if err := writeStaged(stage, clean, append(existing, op.Body...)); err != nil {
			return err
		}
		res.Written = append(res.Written, clean)
	case domain.FileDelete:
		if err := os.RemoveAll(clean); err != nil {
			return fmt.Errorf("catalog: delete %s: %w", clean, err)
		}
		res.Deleted = append(res.Deleted, clean)
	default:
		return fmt.Errorf("catalog: unsupported op %q", op.Op)
	}
	return nil
}

func safeJoin(root, p string) (string, error) {
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	abs := filepath.Clean(p)
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(abs+string(os.PathSeparator), rootAbs+string(os.PathSeparator)) && abs != rootAbs {
		return "", fmt.Errorf("catalog: refusing to write outside project root: %s", abs)
	}
	return abs, nil
}

func writeStaged(stage, finalPath string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return fmt.Errorf("catalog: mkdir %s: %w", filepath.Dir(finalPath), err)
	}
	tmp, err := os.CreateTemp(stage, "op-*")
	if err != nil {
		return fmt.Errorf("catalog: temp: %w", err)
	}
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), finalPath); err != nil {
		return fmt.Errorf("catalog: rename %s: %w", finalPath, err)
	}
	return nil
}

// HashOps returns a stable SHA-256 over the (op,path,body) tuples; recorded
// in installed_capabilities.content_hash so reviewers can spot drift.
func HashOps(ops []domain.FileOp) string {
	h := sha256.New()
	for _, op := range ops {
		_, _ = io.WriteString(h, string(op.Op))
		_, _ = io.WriteString(h, "\x00")
		_, _ = io.WriteString(h, op.Path)
		_, _ = io.WriteString(h, "\x00")
		_, _ = h.Write(op.Body)
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ErrUserDenied is returned by Install when interactive approval is missing.
var ErrUserDenied = errors.New("catalog: approval required (re-run with --yes or accept interactively)")

// ExitCode for use by the cmd layer when ErrUserDenied surfaces.
const ExitCodeUserDenied = constants.ExitUserDeny
