// SPDX-License-Identifier: MIT

package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ropeixoto/harnessx/internal/app/doctor"
)

const boxWidth = 56

func RenderDoctor(w io.Writer, r doctor.Report) {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(boxWidth)

	var b strings.Builder
	b.WriteString(Heading.Render("HarnessX Doctor") + "\n\n")
	b.WriteString(kv("System", fmt.Sprintf("%s %s", r.OS, r.Arch)) + "\n")
	b.WriteString("\n")
	b.WriteString(Heading.Render("Tools") + "\n")
	for _, e := range r.Tools {
		b.WriteString(entryLine(e) + "\n")
	}
	if len(r.LSPs) > 0 {
		b.WriteString("\n")
		b.WriteString(Heading.Render("LSP Servers") + "\n")
		for _, e := range r.LSPs {
			b.WriteString(entryLine(e) + "\n")
		}
	}
	if len(r.Quality) > 0 {
		b.WriteString("\n")
		b.WriteString(Heading.Render("Quality + Supply-chain") + "\n")
		for _, e := range r.Quality {
			b.WriteString(entryLine(e) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(Heading.Render("Agents") + "\n")
	for _, e := range r.Agents {
		b.WriteString(entryLine(e) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(Heading.Render("Project") + "\n")
	b.WriteString(kv("Root", truncate(r.Project.Root, boxWidth-12)) + "\n")
	if r.Project.HarnessReady {
		b.WriteString(kv("Harness DB", Success.Render("✓ ready")) + "\n")
	} else {
		b.WriteString(kv("Harness DB", Muted.Render("not initialised (run `harness init`)")) + "\n")
	}

	fmt.Fprintln(w, box.Render(b.String()))

	hints := collectInstallHints(r)
	if len(hints) > 0 {
		fmt.Fprintln(w, Heading.Render("Recommended installs"))
		for _, h := range hints {
			fmt.Fprintln(w, "  "+Muted.Render("→ harness install "+h))
		}
		if anyLSPMissing(r) {
			fmt.Fprintln(w, "  "+Muted.Render("→ harness install lsp --all   # every LSP at once"))
		}
	}
}

func anyLSPMissing(r doctor.Report) bool {
	for _, e := range r.LSPs {
		if !e.Result.Present || e.Result.Err != nil {
			return true
		}
	}
	return false
}

func collectInstallHints(r doctor.Report) []string {
	var out []string
	seen := map[string]bool{}
	add := func(entries []doctor.Entry) {
		for _, e := range entries {
			if e.Spec.InstallID == "" {
				continue
			}
			if e.Result.Present && e.Result.Err == nil {
				continue
			}
			if !seen[e.Spec.InstallID] {
				seen[e.Spec.InstallID] = true
				out = append(out, e.Spec.InstallID)
			}
		}
	}
	add(r.Tools)
	add(r.LSPs)
	add(r.Quality)
	add(r.Agents)
	return out
}

func entryLine(e doctor.Entry) string {
	label := pad(e.Spec.Label, 14)
	switch {
	case e.Result.Present && e.Result.Err == nil:
		v := e.Result.Version
		if v == "" {
			v = "available"
		}
		return label + Success.Render("✓ ") + Muted.Render(truncate(v, boxWidth-20))
	case e.Result.Present && e.Result.Err != nil:
		return label + Warn.Render("⚠ ") + Muted.Render("present, version probe failed")
	case e.Spec.Required:
		return label + Error.Render("✗ missing (required)")
	default:
		return label + Muted.Render("⚠ not installed")
	}
}

func kv(k, v string) string {
	return pad(k, 14) + v
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return s[:n-1] + "…"
}
