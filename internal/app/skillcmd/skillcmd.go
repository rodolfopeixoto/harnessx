// SPDX-License-Identifier: MIT

// Package skillcmd wires `harness skill list|promote`.
package skillcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/skills"
)

func List(out io.Writer, startDir string) error {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("skill list: db missing (run `harness init`)")
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()
	list, err := skills.List(context.Background(), repo.DB())
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Fprintln(out, "(no skill versions yet)")
		return nil
	}
	fmt.Fprintf(out, "%-26s %3s %8s %8s %s\n", "SKILL", "VER", "SCORE", "ACCEPT", "CREATED")
	for _, s := range list {
		acc := "no"
		if s.Accepted {
			acc = "yes"
		}
		fmt.Fprintf(out, "%-26s %3d %8.2f %8s %s\n",
			s.Name, s.Version, s.Score, acc, s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
	return nil
}

type PromoteOptions struct {
	StartDir string
	Name     string
	File     string
}

func Promote(out io.Writer, opts PromoteOptions) error {
	if opts.Name == "" || opts.File == "" {
		return fmt.Errorf("skill promote: --name and --file are required")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("skill promote: db missing (run `harness init`)")
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()
	b, err := os.ReadFile(opts.File)
	if err != nil {
		return err
	}
	s, err := skills.Promote(context.Background(), repo.DB(), root, opts.Name, string(b), nil)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "promoted skill %s v%d (score %.2f, hash %s)\n",
		s.Name, s.Version, s.Score, s.ContentHash[:12])
	return nil
}
