// SPDX-License-Identifier: MIT

package design

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	hrefRe       = regexp.MustCompile(`(?i)href\s*=\s*"([^"#?]+)"`)
	classRe      = regexp.MustCompile(`(?i)class\s*=\s*"([^"]+)"`)
	dataRouteRe  = regexp.MustCompile(`(?i)data-route\s*=\s*"([^"]+)"`)
	titleRe      = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	cssColorRe   = regexp.MustCompile(`#[0-9A-Fa-f]{3,8}\b`)
	cssSpacingRe = regexp.MustCompile(`\b\d+(?:\.\d+)?(?:px|rem|em)\b`)
	cssVarRe     = regexp.MustCompile(`--([a-zA-Z0-9_\-]+)\s*:\s*([^;]+);`)
	jsHandlerRe  = regexp.MustCompile(`(?i)on(click|submit|change|input|hover|focus|blur)\b`)
)

var assetExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".ico": true, ".woff": true, ".woff2": true, ".ttf": true,
	".otf": true, ".mp4": true, ".webm": true,
}

var componentDirHints = map[string]bool{
	"components": true, "ui": true, "widgets": true, "blocks": true,
}

// Inventory walks the design root and produces a Manifest. It does not
// interpret semantics — that's the job of features.go / roadmap.go.
func Inventory(src Source) (*Manifest, error) {
	m := &Manifest{
		Source:      src.Origin,
		GeneratedAt: time.Now().UTC(),
	}
	colorSet := map[string]struct{}{}
	spacingSet := map[string]struct{}{}
	fontSet := map[string]struct{}{}
	componentSet := map[string]struct{}{}
	interactionSet := map[string]struct{}{}
	flowSet := map[string]struct{}{}

	err := filepath.WalkDir(src.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if name := d.Name(); strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(src.Root, p)
		ext := strings.ToLower(filepath.Ext(rel))

		if assetExts[ext] {
			m.Assets = append(m.Assets, rel)
			return nil
		}
		switch ext {
		case ".html", ".htm":
			page := parsePage(rel, p, interactionSet, flowSet)
			m.Pages = append(m.Pages, page)
			if isComponentPath(rel) {
				componentSet[componentNameFromPath(rel)] = struct{}{}
			}
		case ".css", ".scss", ".sass":
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			for _, c := range cssColorRe.FindAllString(string(b), -1) {
				colorSet[strings.ToLower(c)] = struct{}{}
			}
			for _, s := range cssSpacingRe.FindAllString(string(b), -1) {
				spacingSet[s] = struct{}{}
			}
			for _, v := range cssVarRe.FindAllStringSubmatch(string(b), -1) {
				if strings.HasPrefix(v[1], "font") {
					fontSet[strings.TrimSpace(v[2])] = struct{}{}
				}
			}
		case ".js", ".jsx", ".ts", ".tsx":
			if isComponentPath(rel) {
				componentSet[componentNameFromPath(rel)] = struct{}{}
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			for _, m := range jsHandlerRe.FindAllStringSubmatch(string(b), -1) {
				interactionSet[strings.ToLower(m[1])] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	m.Components = componentsFromSet(componentSet, src.Root)
	m.Styles.Colors = sortedKeys(colorSet)
	m.Styles.Spacing = sortedKeys(spacingSet)
	m.Styles.Fonts = sortedKeys(fontSet)
	m.DetectedFlows = sortedKeys(flowSet)

	// Missing-state heuristics: pages that didn't surface common state
	// classes are flagged for manual review.
	m.MissingStates = missingStatesFromPages(m.Pages, src.Root)
	m.Responsive = responsiveHints(src.Root)

	sort.Strings(m.Assets)
	sort.Slice(m.Pages, func(i, j int) bool { return m.Pages[i].Path < m.Pages[j].Path })

	return m, nil
}

func parsePage(rel, abs string, interactions, flows map[string]struct{}) Page {
	b, _ := os.ReadFile(abs)
	body := string(b)
	id := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	path := pageRoute(rel)
	title := strings.TrimSpace(firstSubmatch(titleRe, body))
	if title == "" {
		title = id
	}
	page := Page{ID: id, Path: path, Title: title, File: rel}

	// Components used via class="" (very rough — Phase 7 doesn't ship a
	// full DOM parser; the heuristic surfaces obvious component reuse).
	classes := map[string]bool{}
	for _, m := range classRe.FindAllStringSubmatch(body, -1) {
		for _, c := range strings.Fields(m[1]) {
			if looksLikeComponentClass(c) {
				classes[c] = true
			}
		}
	}
	for c := range classes {
		page.Components = append(page.Components, c)
	}
	sort.Strings(page.Components)

	// Interactions from inline JS handlers and data-route hops.
	pageActions := map[string]struct{}{}
	for _, m := range jsHandlerRe.FindAllStringSubmatch(body, -1) {
		action := strings.ToLower(m[1])
		pageActions[action] = struct{}{}
		interactions[action] = struct{}{}
	}
	for _, m := range dataRouteRe.FindAllStringSubmatch(body, -1) {
		flows[id+" → "+m[1]] = struct{}{}
		pageActions["navigate:"+m[1]] = struct{}{}
	}
	for _, m := range hrefRe.FindAllStringSubmatch(body, -1) {
		target := m[1]
		if strings.HasPrefix(target, "http") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		flows[id+" → "+target] = struct{}{}
	}
	for a := range pageActions {
		page.Interactions = append(page.Interactions, a)
	}
	sort.Strings(page.Interactions)

	return page
}

func pageRoute(rel string) string {
	rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")
	rel = strings.TrimSuffix(rel, ".html")
	rel = strings.TrimSuffix(rel, ".htm")
	if rel == "index" || strings.HasSuffix(rel, "/index") {
		return "/" + strings.TrimSuffix(rel, "index")
	}
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return rel
}

func looksLikeComponentClass(c string) bool {
	if len(c) < 3 {
		return false
	}
	// PascalCase or kebab-case-with-prefix
	first := c[0]
	return (first >= 'A' && first <= 'Z') ||
		strings.HasPrefix(c, "ui-") || strings.HasPrefix(c, "c-") || strings.HasPrefix(c, "comp-")
}

func isComponentPath(rel string) bool {
	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		if componentDirHints[strings.ToLower(p)] {
			return true
		}
	}
	return false
}

func componentNameFromPath(rel string) string {
	base := filepath.Base(rel)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func componentsFromSet(set map[string]struct{}, root string) []Component {
	keys := sortedKeys(set)
	out := make([]Component, 0, len(keys))
	for _, k := range keys {
		out = append(out, Component{Name: k, File: findFileFor(root, k)})
	}
	return out
}

func findFileFor(root, name string) string {
	var found string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())), name) {
			rel, _ := filepath.Rel(root, p)
			found = rel
		}
		return nil
	})
	return found
}

func missingStatesFromPages(pages []Page, root string) []string {
	stateClasses := []string{"loading", "empty", "error", "disabled"}
	missing := map[string]struct{}{}
	for _, page := range pages {
		body, _ := os.ReadFile(filepath.Join(root, page.File))
		text := strings.ToLower(string(body))
		for _, state := range stateClasses {
			if !strings.Contains(text, state) {
				missing[page.ID+":"+state] = struct{}{}
			}
		}
	}
	return sortedKeys(missing)
}

func responsiveHints(root string) []string {
	var notes []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".css" && ext != ".html" {
			return nil
		}
		b, _ := os.ReadFile(p)
		if strings.Contains(string(b), "@media") {
			rel, _ := filepath.Rel(root, p)
			notes = append(notes, rel+":@media")
		}
		return nil
	})
	if len(notes) == 0 {
		notes = []string{"no @media queries detected — assume mobile parity needs new work"}
	}
	return notes
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func firstSubmatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
