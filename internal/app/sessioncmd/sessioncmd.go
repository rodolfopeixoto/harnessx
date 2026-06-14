// SPDX-License-Identifier: MIT

// Package sessioncmd wires `harness session show <id>` — read-only view
// of one session's runs + sensors + cost from sqlite.
package sessioncmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type ShowOptions struct {
	StartDir string
	ID       string
}

func Show(out io.Writer, opts ShowOptions) error {
	if opts.ID == "" {
		return fmt.Errorf("session show: missing id")
	}
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("session: db missing (run `harness init`)")
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	ctx := context.Background()
	row := repo.DB().QueryRowContext(ctx, `
		select id, project_path, mode, status, started_at, finished_at,
		       total_cost_usd, total_input_tokens, total_output_tokens
		from sessions where id = ?`, opts.ID)
	var sessID, project, mode, status, started string
	var finished any
	var cost float64
	var inTok, outTok int64
	if err := row.Scan(&sessID, &project, &mode, &status, &started, &finished, &cost, &inTok, &outTok); err != nil {
		return fmt.Errorf("session: %w", err)
	}
	fmt.Fprintf(out, "session %s\n", sessID)
	fmt.Fprintf(out, "  project: %s\n", project)
	fmt.Fprintf(out, "  mode:    %s\n", mode)
	fmt.Fprintf(out, "  status:  %s\n", status)
	fmt.Fprintf(out, "  started: %s\n", started)
	if s, ok := finished.(string); ok && s != "" {
		fmt.Fprintf(out, "  finished: %s\n", s)
	}
	fmt.Fprintf(out, "  total cost: $%.4f, tokens %d/%d\n", cost, inTok, outTok)

	fmt.Fprintln(out, "\nruns:")
	rows, err := repo.DB().QueryContext(ctx, `
		select id, stage, agent, status, latency_ms, exit_code,
		       input_tokens, output_tokens, estimated_cost_usd, fallback_from, error_type
		from runs where session_id = ? order by started_at`, opts.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var n int
	for rows.Next() {
		var runID, stage, st string
		var agent, fallback, errType any
		var latency, inT, outT, exit any
		var costR any
		if err := rows.Scan(&runID, &stage, &agent, &st, &latency, &exit, &inT, &outT, &costR, &fallback, &errType); err != nil {
			return err
		}
		fmt.Fprintf(out, "  - %s [%s] %s agent=%v latency=%vms exit=%v tokens=%v/%v cost=%v fallback_from=%v err=%v\n",
			runID, st, stage, agent, latency, exit, inT, outT, costR, fallback, errType)
		n++

		srows, _ := repo.DB().QueryContext(ctx, `
			select sensor, status, duration_ms from sensor_results where run_id = ? order by id`, runID)
		for srows != nil && srows.Next() {
			var sensor, ss string
			var dur int64
			_ = srows.Scan(&sensor, &ss, &dur)
			fmt.Fprintf(out, "      sensor %s [%s] %dms\n", sensor, ss, dur)
		}
		if srows != nil {
			srows.Close()
		}
	}
	if n == 0 {
		fmt.Fprintln(out, "  (no runs recorded)")
	}
	return nil
}
