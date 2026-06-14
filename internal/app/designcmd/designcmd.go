// SPDX-License-Identifier: MIT

// Package designcmd wires `harness design-to-product` on top of internal/design.
package designcmd

import (
	stdctx "context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/design"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type Options struct {
	StartDir string
	Prompt   string
	Source   string // explicit --source ZIP/folder; empty = resolve from prompt
}

func Run(ctx stdctx.Context, opts Options, out io.Writer) (*design.BuildResult, error) {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return nil, err
	}
	src := opts.Source
	if src == "" {
		src = resolveFromPrompt(opts.Prompt, root)
	}
	if src == "" {
		return nil, errors.New("design-to-product: no design source resolved (pass --source <zip|folder> or include a path in the prompt)")
	}
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("design-to-product: source %s: %w", src, err)
	}

	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)

	var repo *sqlite.Repo
	var sess domain.Session
	var run domain.Run
	if _, err := os.Stat(dbPath); err == nil {
		repo, err = sqlite.Open(dbPath)
		if err == nil {
			defer repo.Close()
			now := time.Now().UTC()
			sess = domain.Session{
				ID: ids.New(), ProjectPath: root, Mode: domain.ModeDesignToProduct,
				Status: domain.StatusRunning, StartedAt: now,
			}
			run = domain.Run{
				ID: ids.New(), SessionID: sess.ID, Stage: domain.Stage("design_to_product"),
				Status: domain.StatusRunning, StartedAt: now,
			}
			_ = repo.CreateSession(ctx, sess)
			_ = repo.CreateRun(ctx, run)
		}
	}

	fmt.Fprintf(out, "Detected Design-to-Product mode (source=%s)\n", src)
	fmt.Fprintln(out, "Stages: resolve → extract → inventory → manifest → feature map → roadmap → API contracts → flow map → image cache")

	res, err := design.Build(design.BuildOptions{Root: root, Source: src})
	if err != nil {
		if repo != nil {
			end := time.Now().UTC()
			_ = repo.FinishRun(ctx, run.ID, domain.StatusFailed, end, 1)
			_ = repo.FinishSession(ctx, sess.ID, domain.StatusFailed, end)
		}
		return nil, err
	}

	fmt.Fprintln(out, "Product maps written:")
	fmt.Fprintf(out, "  - %s\n", res.ManifestPath)
	fmt.Fprintf(out, "  - %s\n", res.FeatureMapPath)
	fmt.Fprintf(out, "  - %s\n", res.ToggleMapPath)
	fmt.Fprintf(out, "  - %s\n", res.RoadmapPath)
	fmt.Fprintf(out, "  - %s\n", res.APIContractsPath)
	fmt.Fprintf(out, "  - %s\n", res.FlowMapPath)
	fmt.Fprintf(out, "Images analysed: %d\n", res.ImagesAnalyzed)
	if res.Manifest != nil {
		fmt.Fprintf(out, "Summary: %d pages, %d components, %d assets, %d flows\n",
			len(res.Manifest.Pages), len(res.Manifest.Components),
			len(res.Manifest.Assets), len(res.Manifest.DetectedFlows))
	}

	if repo != nil {
		end := time.Now().UTC()
		_ = repo.FinishRun(ctx, run.ID, domain.StatusSucceeded, end, 0)
		_ = repo.FinishSession(ctx, sess.ID, domain.StatusSucceeded, end)
	}
	return res, nil
}

// resolveFromPrompt mines the prompt for a path-shaped token. Matches
// .zip files and existing directories. Returns empty when no candidate.
func resolveFromPrompt(prompt, root string) string {
	for _, tok := range strings.Fields(prompt) {
		t := strings.Trim(tok, "\"'`,;()")
		if t == "" {
			continue
		}
		if !strings.ContainsAny(t, "/.") {
			continue
		}
		// expand relative to project root
		cand := t
		if !filepath.IsAbs(t) {
			cand = filepath.Join(root, t)
		}
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return ""
}
