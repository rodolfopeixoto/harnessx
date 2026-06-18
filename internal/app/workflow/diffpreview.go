// SPDX-License-Identifier: MIT

package workflow

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ropeixoto/harnessx/internal/ui"
)

func printDiffPreview(out io.Writer, diffPath, statPath string) {
	if statPath != "" {
		if body, err := os.ReadFile(statPath); err == nil && len(body) > 0 {
			fmt.Fprint(out, ui.Muted.Render("  ┌── stat ──────────────────────────────\n"))
			for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
				fmt.Fprintf(out, "  %s %s\n", ui.Muted.Render("│"), line)
			}
			fmt.Fprint(out, ui.Muted.Render("  └────────────────────────────────────────\n"))
		}
	}
	f, err := os.Open(diffPath)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprint(out, ui.Muted.Render("  ┌── diff (first 40 lines) ─────────────\n"))
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() && n < 40 {
		line := sc.Text()
		colour := line
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			colour = ui.Success.Render(line)
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			colour = ui.Error.Render(line)
		case strings.HasPrefix(line, "@@"):
			colour = ui.Info.Render(line)
		case strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			colour = ui.Heading.Render(line)
		}
		fmt.Fprintf(out, "  %s %s\n", ui.Muted.Render("│"), colour)
		n++
	}
	if n == 40 {
		fmt.Fprintf(out, "  %s %s\n", ui.Muted.Render("│"), ui.Muted.Render("... (truncated; cat the file for the rest)"))
	}
	fmt.Fprint(out, ui.Muted.Render("  └────────────────────────────────────────\n"))
}
