// SPDX-License-Identifier: MIT

// Package dashboardcmd wires `harness dashboard`.
package dashboardcmd

import (
	stdctx "context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	dashboardhttp "github.com/ropeixoto/harnessx/internal/adapters/http"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type Options struct {
	StartDir string
	Addr     string
	Open     bool
}

func Run(ctx stdctx.Context, opts Options, out io.Writer) error {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return err
	}
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:7373"
	}
	dist := candidateDist(root)
	srv, err := dashboardhttp.New(dashboardhttp.Options{
		Root: root, Addr: opts.Addr, Dist: dist,
	})
	if err != nil {
		return err
	}

	url := "http://" + opts.Addr
	fmt.Fprintf(out, "harness: dashboard serving on %s\n", url)
	if dist != "" {
		fmt.Fprintf(out, "  serving React build from %s\n", dist)
	} else {
		fmt.Fprintln(out, "  serving built-in HTML (run `make dashboard-build` for React UI)")
	}
	if opts.Open {
		go func() {
			// Give the listener a moment to bind before opening.
			time.Sleep(250 * time.Millisecond)
			_ = openBrowser(url)
		}()
	}

	if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func candidateDist(root string) string {
	candidates := []string{
		filepath.Join(root, "web", "dashboard", "dist"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	}
	return nil
}
