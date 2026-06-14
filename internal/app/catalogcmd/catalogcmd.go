// SPDX-License-Identifier: MIT

// Package catalogcmd is the thin renderer behind `harness catalog ...`.
package catalogcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/ropeixoto/harnessx/internal/catalog"
	"github.com/ropeixoto/harnessx/internal/catalog/kinds"
	"github.com/ropeixoto/harnessx/internal/domain"
)

// New returns a Catalog with every bundled kind registered.
func New() *catalog.Catalog {
	c := catalog.New()
	for _, k := range kinds.All() {
		c.Register(k)
	}
	return c
}

// List walks every registered kind and prints a unified table.
func List(ctx context.Context, root string, kind domain.CapabilityKind, out io.Writer) error {
	c := New()
	var caps []domain.Capability
	var err error
	if kind != "" {
		caps, err = c.DiscoverKind(ctx, root, kind)
	} else {
		caps, err = c.Discover(ctx, root)
	}
	if err != nil {
		return err
	}
	if len(caps) == 0 {
		fmt.Fprintln(out, "no capabilities discovered")
		return nil
	}
	sort.SliceStable(caps, func(i, j int) bool {
		if caps[i].Kind != caps[j].Kind {
			return caps[i].Kind < caps[j].Kind
		}
		return caps[i].Name < caps[j].Name
	})
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KIND\tNAME\tSTATUS\tSOURCE\tVERSION\tDESCRIPTION")
	for _, cap := range caps {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			cap.Kind, cap.Name, cap.Status, cap.Source, cap.Version, shorten(cap.Description, 64))
	}
	return tw.Flush()
}

// Show prints the full record for one capability.
func Show(ctx context.Context, root string, kind domain.CapabilityKind, name string, out io.Writer) error {
	cap, err := New().Show(ctx, root, kind, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s/%s\n  status:  %s\n  source:  %s\n  version: %s\n  scope:   %s\n  config:  %s\n  manifest:%s\n  %s\n",
		cap.Kind, cap.Name, cap.Status, cap.Source, cap.Version, cap.Scope, cap.ConfigPath, cap.ManifestPath, cap.Description)
	return nil
}

// Plan renders the diff that Install would apply.
func Plan(ctx context.Context, root string, kind domain.CapabilityKind, name string, out io.Writer) error {
	c := New()
	ops, err := c.Plan(ctx, root, kind, name)
	if err != nil {
		return err
	}
	return printDiff(out, ops)
}

// Install applies the plan after approval.
func Install(ctx context.Context, root string, kind domain.CapabilityKind, name string, yes bool, dryRun bool, in io.Reader, out io.Writer) error {
	c := New()
	ops, err := c.Plan(ctx, root, kind, name)
	if err != nil {
		return err
	}
	if err := printDiff(out, ops); err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintln(out, "(dry-run; no files written)")
		return nil
	}
	if !yes {
		if !confirm(in, out) {
			return catalog.ErrUserDenied
		}
	}
	res, err := catalog.Apply(ctx, root, ops)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "✓ installed %s/%s (%d files; hash=%s)\n", kind, name, len(res.Written), catalog.HashOps(ops)[:12])
	return nil
}

// Remove deletes the installed config for a capability.
func Remove(ctx context.Context, root string, kind domain.CapabilityKind, name string, out io.Writer) error {
	cap, err := New().Show(ctx, root, kind, name)
	if err != nil {
		if errors.Is(err, catalog.ErrUnknownCapability) {
			return fmt.Errorf("catalog: nothing to remove (%s/%s not found)", kind, name)
		}
		return err
	}
	if cap.ConfigPath == "" {
		return fmt.Errorf("catalog: %s/%s has no installed config", kind, name)
	}
	if err := os.Remove(cap.ConfigPath); err != nil {
		return err
	}
	fmt.Fprintf(out, "✓ removed %s/%s (config %s)\n", kind, name, cap.ConfigPath)
	return nil
}

func printDiff(out io.Writer, ops []domain.FileOp) error {
	var buf bytes.Buffer
	for _, op := range ops {
		fmt.Fprintf(&buf, "  %s %s\n", op.Op, op.Path)
		if op.Op == domain.FileCreate || op.Op == domain.FileOverwrite {
			fmt.Fprintln(&buf, indent(string(op.Body), "    │ "))
		}
	}
	_, err := io.Copy(out, &buf)
	return err
}

func confirm(in io.Reader, out io.Writer) bool {
	if in == nil {
		return false
	}
	fmt.Fprint(out, "apply? [y/N] ")
	buf := make([]byte, 4)
	n, err := in.Read(buf)
	if err != nil || n == 0 {
		fmt.Fprintln(out)
		return false
	}
	c := buf[0]
	fmt.Fprintln(out)
	return c == 'y' || c == 'Y'
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func indent(s, prefix string) string {
	if s == "" {
		return ""
	}
	var b bytes.Buffer
	for i, line := range splitLines(s) {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(prefix)
		b.WriteString(line)
	}
	return b.String()
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
