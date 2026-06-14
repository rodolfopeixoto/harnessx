// SPDX-License-Identifier: MIT

// Package routescmd wires `harness routes` — read-only inspection of the
// task → agent-chain map. Combines bundled defaults + user overrides from
// .harness/config/routes.yaml and prints the resolved Selection per task.
package routescmd

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/ropeixoto/harnessx/internal/app/agentcmd"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
	"github.com/ropeixoto/harnessx/internal/router"
)

type Options struct {
	StartDir string
	Task     string // empty = show every route
}

var defaultRoutes = router.Defaults

func Run(out io.Writer, opts Options) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	reg, _, err := agentcmd.LoadAll(root)
	if err != nil {
		return err
	}
	routes := defaultRoutes(reg)
	source := "bundled"
	if user, err := router.LoadConfig(filepath.Join(root, ".harness", "config", "routes.yaml")); err == nil && user != nil {
		for k, v := range user {
			routes[k] = v
		}
		source = "bundled + " + filepath.Join(".harness", "config", "routes.yaml")
	}

	r := router.New(reg, routes)
	keys := make([]string, 0, len(routes))
	for k := range routes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if opts.Task != "" {
		keys = []string{opts.Task}
	}

	fmt.Fprintf(out, "Routes source: %s\n", source)
	fmt.Fprintf(out, "Registered agents: %v\n\n", reg.IDs())

	for _, task := range keys {
		cfg, ok := routes[task]
		if !ok {
			fmt.Fprintf(out, "%-22s — no route configured\n", task)
			continue
		}
		d, err := r.Select(task)
		if err != nil {
			fmt.Fprintf(out, "%-22s ✗ %v\n", task, err)
			continue
		}
		fmt.Fprintf(out, "%-22s budget=$%.2f primary=%s fallback=%v\n",
			task, cfg.BudgetUSD, cfg.Primary, cfg.Fallback)
		fmt.Fprintf(out, "  resolved chain: ")
		for i, a := range d.Chain {
			if i > 0 {
				fmt.Fprint(out, " → ")
			}
			fmt.Fprint(out, a.ID())
		}
		fmt.Fprintln(out)
		for _, reason := range d.Reasons[1:] {
			fmt.Fprintf(out, "  reason: %s\n", reason)
		}
		fmt.Fprintln(out)
	}
	return nil
}
