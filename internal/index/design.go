// SPDX-License-Identifier: MIT

package index

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	tailwindColorRe   = regexp.MustCompile(`#[0-9A-Fa-f]{3,8}\b`)
	tailwindSpacingRe = regexp.MustCompile(`\b\d+(?:\.\d+)?(?:px|rem|em)\b`)
	tailwindFontRe    = regexp.MustCompile(`(?i)fontFamily\s*:\s*\{([^}]*)\}`)
	tailwindScreenRe  = regexp.MustCompile(`(?i)screens\s*:\s*\{([^}]*)\}`)
)

func BuildDesignSystem(root string) DesignSystem {
	d := DesignSystem{Note: "no design tokens detected yet"}
	candidates := []string{
		"tailwind.config.js", "tailwind.config.ts",
		"design-tokens.json", "tokens.json",
		filepath.Join(".harness", "product", "design-manifest.json"),
	}
	colors := map[string]struct{}{}
	spacing := map[string]struct{}{}
	fonts := map[string]struct{}{}
	breakpoints := map[string]struct{}{}
	for _, c := range candidates {
		full := filepath.Join(root, c)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		d.Detected = true
		d.Sources = append(d.Sources, c)
		if strings.HasPrefix(c, "tailwind.config") {
			b, _ := os.ReadFile(full)
			extractTokens(string(b), colors, spacing, fonts, breakpoints)
		}
		if c == "design-tokens.json" || c == "tokens.json" || strings.HasSuffix(c, "design-manifest.json") {
			b, _ := os.ReadFile(full)
			for _, c := range tailwindColorRe.FindAllString(string(b), -1) {
				colors[strings.ToLower(c)] = struct{}{}
			}
			for _, s := range tailwindSpacingRe.FindAllString(string(b), -1) {
				spacing[s] = struct{}{}
			}
		}
	}
	if d.Detected {
		d.Note = ""
		d.Colors = sortedSet(colors)
		d.Spacing = sortedSet(spacing)
		d.Fonts = sortedSet(fonts)
		d.Breakpoints = sortedSet(breakpoints)
	}
	return d
}

func extractTokens(body string, colors, spacing, fonts, breakpoints map[string]struct{}) {
	for _, c := range tailwindColorRe.FindAllString(body, -1) {
		colors[strings.ToLower(c)] = struct{}{}
	}
	for _, s := range tailwindSpacingRe.FindAllString(body, -1) {
		spacing[s] = struct{}{}
	}
	if m := tailwindFontRe.FindStringSubmatch(body); len(m) == 2 {
		for _, line := range strings.Split(m[1], ",") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if idx := strings.Index(line, ":"); idx > 0 {
				name := strings.Trim(line[:idx], "\"' ")
				if name != "" {
					fonts[name] = struct{}{}
				}
			}
		}
	}
	if m := tailwindScreenRe.FindStringSubmatch(body); len(m) == 2 {
		for _, line := range strings.Split(m[1], ",") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if idx := strings.Index(line, ":"); idx > 0 {
				name := strings.Trim(line[:idx], "\"' ")
				if name != "" {
					breakpoints[name] = struct{}{}
				}
			}
		}
	}
}

func sortedSet(s map[string]struct{}) []string {
	if len(s) == 0 {
		return nil
	}
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
