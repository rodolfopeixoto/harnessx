// SPDX-License-Identifier: MIT

// Package memorycmd wires `harness memory list|promote` on top of
// internal/memory. Promotion is evidence-gated; List is read-only.
package memorycmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/memory"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/i18n"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type ListOptions struct {
	StartDir string
	Limit    int
	Scope    string // optional filter
}

func List(out io.Writer, opts ListOptions) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		fmt.Fprintln(out, "memory: .harness not initialised (run `harness init` first)")
		return nil
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	q := `select id, scope, kind, content, evidence_run_id, confidence, updated_at
	      from memories`
	args := []any{}
	if opts.Scope != "" {
		q += ` where scope = ?`
		args = append(args, opts.Scope)
	}
	q += ` order by updated_at desc limit ?`
	args = append(args, limit)

	rows, err := repo.DB().QueryContext(context.Background(), q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Fprintf(out, "%-26s %-10s %-15s %-5s %s\n", "ID", "SCOPE", "KIND", "CONF", "CONTENT")
	count := 0
	for rows.Next() {
		var id, scope, kind, content, evidence, updated string
		var conf float64
		if err := rows.Scan(&id, &scope, &kind, &content, &evidence, &conf, &updated); err != nil {
			return err
		}
		fmt.Fprintf(out, "%-26s %-10s %-15s %.2f  %s\n", id[:min(len(id), 26)], scope, kind, conf, truncate(content, 80))
		count++
	}
	if count == 0 {
		fmt.Fprintln(out, i18n.T("memory.empty"))
	}
	return nil
}

type PromoteOptions struct {
	StartDir   string
	Scope      string
	Kind       string
	Content    string
	RunID      string
	Confidence float64
}

func Promote(ctx context.Context, out io.Writer, opts PromoteOptions) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("memory: .harness not initialised — run `harness init`")
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	adapter := sqlAdapter{repo: repo}
	m, err := memory.Promote(ctx, repo, memory.Candidate{
		Scope:         opts.Scope,
		Kind:          opts.Kind,
		Content:       opts.Content,
		EvidenceRunID: opts.RunID,
		Confidence:    opts.Confidence,
	}, adapter)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s %s (scope=%s kind=%s confidence=%.2f)\n",
		i18n.T("memory.promoted"), m.ID, m.Scope, m.Kind, m.Confidence)
	return nil
}

type sqlAdapter struct{ repo *sqlite.Repo }

func (a sqlAdapter) ExecContext(ctx context.Context, q string, args ...any) (any, error) {
	return a.repo.DB().ExecContext(ctx, q, args...)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// Now is exported so tests can stub.
var Now = time.Now
