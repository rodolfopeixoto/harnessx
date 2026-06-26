// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/ui"
)

func runShipWatch(ctx context.Context, out io.Writer, opts shipOptions) error {
	root, err := cwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s ship --watch on %s (interval=%s)\n", ui.MarkInfo(), root, opts.watchInterval)
	last := ""
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		hash, err := hashProjectFiles(root)
		if err != nil {
			fmt.Fprintf(out, "%s watch hash err: %v\n", ui.MarkWarn(), err)
		}
		if hash != last {
			last = hash
			fmt.Fprintf(out, "%s change detected — running ship\n", ui.MarkInfo())
			if err := runShip(ctx, out, opts); err != nil {
				fmt.Fprintf(out, "%s ship: %v\n", ui.MarkWarn(), err)
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(opts.watchInterval):
		}
	}
}

var watchSkipDirs = map[string]bool{
	".git": true, ".harness": true, "node_modules": true, "vendor": true,
	"target": true, "dist": true, "build": true, ".venv": true, "venv": true,
	"__pycache__": true, ".cache": true,
}

func hashProjectFiles(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if watchSkipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				if p == root {
					return nil
				}
				return filepath.SkipDir
			}
			return nil
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil
		}
		fmt.Fprintf(h, "%s|%d|%d\n", p, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
