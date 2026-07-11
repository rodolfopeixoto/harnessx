// SPDX-License-Identifier: MIT

// Package agentcmd implements the `harness agent …` subcommands.
package agentcmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/certify"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

// List renders the resolved adapter table with certification + experimental
// state pulled from the local project SQLite (when present).
func List(out io.Writer, startDir string) error {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	reg, sources, err := LoadAll(root)
	if err != nil {
		return err
	}
	if len(reg.IDs()) == 0 {
		fmt.Fprintln(out, "no agent adapters registered")
		return nil
	}
	experimentals := loadExperimentalFlags(root)
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	var repo *sqlite.Repo
	if _, err := os.Stat(config.Resolve(root, cfg.Database.Path)); err == nil {
		repo, _ = sqlite.Open(config.Resolve(root, cfg.Database.Path))
		if repo != nil {
			defer repo.Close()
		}
	}
	fmt.Fprintf(out, "%-12s %-22s %-10s %-4s %s\n", "ID", "NAME", "CERT", "EXP", "SOURCE")
	for _, id := range reg.IDs() {
		a, _ := reg.Get(id)
		cert := "—"
		if repo != nil {
			if c, err := repo.LatestAgentCertification(context.Background(), id); err == nil {
				cert = fmt.Sprintf("%s/%d", c.Status, c.Score)
			}
		}
		exp := " "
		if experimentals[id] {
			exp = "★"
		}
		fmt.Fprintf(out, "%-12s %-22s %-10s %-4s %s\n", a.ID(), truncate(a.Name(), 22), cert, exp, sources[id])
	}
	return nil
}

// CertifyOptions bundles adapter certification inputs; StartDir controls
// project resolution and Override lets tests inject a fake adapter.
type CertifyOptions struct {
	ID            string
	SkipRun       bool
	SimpleTimeout time.Duration
	StartDir      string
	Override      agents.AgentAdapter
}

func Certify(ctx context.Context, out io.Writer, opts CertifyOptions) (certify.Result, error) {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return certify.Result{}, err
	}
	var a agents.AgentAdapter
	if opts.Override != nil {
		a = opts.Override
	} else {
		reg, _, err := LoadAll(root)
		if err != nil {
			return certify.Result{}, err
		}
		var ok bool
		a, ok = reg.Get(opts.ID)
		if !ok {
			return certify.Result{}, fmt.Errorf("agent certify: %q not registered", opts.ID)
		}
	}

	res := certify.Run(ctx, a, certify.Options{SkipRun: opts.SkipRun, SimpleTimeout: opts.SimpleTimeout})

	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err == nil {
		repo, err := sqlite.Open(dbPath)
		if err == nil {
			defer repo.Close()
			_ = repo.WriteAgentCertification(ctx, domain.AgentCertification{
				ID: ids.New(), AgentID: a.ID(), CLIVersion: res.CLIVersion,
				AdapterVersion: "1", Score: res.Score, Status: res.Status,
				DetailsJSON: res.DetailsJSON(), CertifiedAt: time.Now().UTC(),
			})
		}
	}

	renderCertification(out, a, res)
	return res, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
