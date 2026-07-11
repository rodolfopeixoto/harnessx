// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	osexec "os/exec"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/projectcfg"
	"github.com/ropeixoto/harnessx/internal/scaffoldpkg"
	"github.com/ropeixoto/harnessx/internal/venvinstall"
)

func applyNewScaffold(abs string, opts *newOptions, out io.Writer) error {
	m, err := scaffoldpkg.Load(opts.stack)
	if err != nil {
		return err
	}
	name := opts.name
	if name == "" {
		name = filepath.Base(abs)
	}
	res, err := scaffoldpkg.Apply(m, scaffoldpkg.ApplyOptions{Root: abs, Name: name})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "new: scaffold %s applied — %d files\n", opts.stack, len(res.Created))
	cfg := projectcfg.FromMeta(m.Language, map[string]string{
		"lint": m.LintCommand,
		"test": m.TestCommand,
		"run":  m.RunCommand,
		"dev":  m.RunCommand,
	})
	if err := projectcfg.Save(abs, cfg); err != nil {
		fmt.Fprintf(out, "new: warning project.yaml: %v\n", err)
	}
	if opts.withDeps {
		runPostStepsInDir(out, abs, m)
	}
	return nil
}

func runPostStepsInDir(out io.Writer, root string, m scaffoldpkg.Meta) {
	if m.Language == "python" || m.Language == "python-ecommerce" {
		res, err := venvinstall.Install(context.Background(), root, "requirements.txt", out)
		if err != nil {
			fmt.Fprintf(out, "  ✗ venv install failed across every strategy: %v\n", err)
			fmt.Fprintln(out, "    fix: install uv (https://docs.astral.sh/uv/) or python3.11/3.12/3.13 and rerun --with-deps")
			return
		}
		fmt.Fprintf(out, "  ✓ deps installed via %s strategy\n", res.Strategy)
		return
	}
	for _, step := range m.PostSteps {
		fmt.Fprintf(out, "new: post-step %s — %v\n", step.Name, step.Cmd)
		cmd := osexec.Command(step.Cmd[0], step.Cmd[1:]...)
		cmd.Dir = root
		cmd.Stdout = out
		cmd.Stderr = out
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(out, "  ✗ %s failed: %v\n", step.Name, err)
		}
	}
}
