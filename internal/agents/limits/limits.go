// SPDX-License-Identifier: MIT

// Package limits exposes per-adapter capability ceilings so the
// caller can sanitise local skill descriptions before spawning the
// upstream CLI. Each entry is conservative: actual provider limits
// may be higher, but enforcing the documented floor avoids
// upstream-CLI crashes when long descriptions are submitted.
package limits

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Adapter holds the limits HarnessX honours when preparing the local
// skill bundle for a specific adapter.
type Adapter struct {
	MaxSkillDescriptionChars int
}

// Catalog is the documented limit table. Extend conservatively —
// changes here gate the warning + truncation flow in PrepareSkills.
var Catalog = map[string]Adapter{
	"codex":  {MaxSkillDescriptionChars: 1024},
	"claude": {MaxSkillDescriptionChars: 8192},
}

// ForAdapter returns the documented limits for the adapter id, or
// an empty Adapter when none registered (caller treats zero as "no
// limit known").
func ForAdapter(id string) Adapter {
	return Catalog[id]
}

// SkillReport records what PrepareSkills saw for one skill file.
type SkillReport struct {
	Path            string
	DescLen         int
	Truncated       bool
	SanitisedPath   string
	SanitisedReason string
}

// PrepareSkills scans every SKILL.md the upstream CLI would
// auto-load (well-known dirs given by `roots`) and copies each
// over-budget file into a per-session skills directory under
// `outDir` with the `description:` field truncated. Returns the
// list of reports so the caller can WARN once per session per
// path. Untouched files are not reported.
//
// The function always returns a non-nil sanitiseDir and a
// CODEX_SKIP_SKILLS-friendly comma list of paths that could not be
// rewritten (e.g. permission denied) — callers expose that via env
// so the upstream CLI skips them entirely.
func PrepareSkills(adapterID string, roots []string, outDir string) (sanitiseDir string, reports []SkillReport, skipPaths []string, err error) {
	cap := ForAdapter(adapterID).MaxSkillDescriptionChars
	if cap <= 0 {
		return "", nil, nil, nil
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", nil, nil, err
	}
	for _, root := range roots {
		entries, _ := os.ReadDir(root)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(root, e.Name(), "SKILL.md")
			info, ierr := os.Stat(path)
			if ierr != nil || info.IsDir() {
				continue
			}
			descLen, truncated, sanitised, serr := sanitiseSkill(path, outDir, e.Name(), cap)
			if serr != nil {
				skipPaths = append(skipPaths, path)
				reports = append(reports, SkillReport{Path: path, DescLen: descLen, Truncated: false, SanitisedReason: serr.Error()})
				continue
			}
			if !truncated {
				continue
			}
			reports = append(reports, SkillReport{
				Path:          path,
				DescLen:       descLen,
				Truncated:     true,
				SanitisedPath: sanitised,
			})
		}
	}
	return outDir, reports, skipPaths, nil
}

func sanitiseSkill(srcPath, outDir, name string, cap int) (descLen int, truncated bool, sanitised string, err error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return 0, false, "", err
	}
	defer f.Close()

	var head []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	tail := []string{}
	descFound := false
	var descRaw strings.Builder
	state := "preamble" // preamble → description → body
	for scanner.Scan() {
		line := scanner.Text()
		switch state {
		case "preamble":
			head = append(head, line)
			if t := strings.TrimSpace(line); strings.HasPrefix(t, "description:") {
				descFound = true
				rest := strings.TrimSpace(strings.TrimPrefix(t, "description:"))
				descRaw.WriteString(rest)
				if !strings.HasSuffix(t, "|") && !strings.HasSuffix(t, ">") {
					state = "body"
				} else {
					state = "description"
				}
			}
		case "description":
			// continuation of YAML multi-line scalar — gather until
			// a non-indented line shows up
			if line != "" && (line[0] == ' ' || line[0] == '\t') {
				descRaw.WriteString("\n")
				descRaw.WriteString(strings.TrimLeft(line, " \t"))
				continue
			}
			state = "body"
			tail = append(tail, line)
		case "body":
			tail = append(tail, line)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return 0, false, "", scanErr
	}
	desc := descRaw.String()
	descLen = len(desc)
	if !descFound || descLen <= cap {
		return descLen, false, "", nil
	}
	trimmed := desc[:cap-1] + "…"

	dst := filepath.Join(outDir, name)
	if mkErr := os.MkdirAll(dst, 0o755); mkErr != nil {
		return descLen, false, "", mkErr
	}
	out, ferr := os.Create(filepath.Join(dst, "SKILL.md"))
	if ferr != nil {
		return descLen, false, "", ferr
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	for _, line := range head {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "description:") {
			fmt.Fprintf(w, "description: %s\n", trimmed)
		} else {
			fmt.Fprintln(w, line)
		}
	}
	for _, line := range tail {
		fmt.Fprintln(w, line)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return descLen, false, "", flushErr
	}
	return descLen, true, filepath.Join(dst, "SKILL.md"), nil
}

// WriteReport emits a single WARN line per skill that was either
// truncated or skipped. Caller passes their session-scoped writer
// (e.g. live REPL output) so the warning lands in front of the user
// exactly once.
var reportedOnce sync.Map

func WriteReport(w io.Writer, sessionID string, reports []SkillReport) {
	for _, r := range reports {
		key := sessionID + "|" + r.Path
		if _, seen := reportedOnce.LoadOrStore(key, struct{}{}); seen {
			continue
		}
		if r.Truncated {
			fmt.Fprintf(w, "  WARN skill description %d chars > limit, truncated: %s\n", r.DescLen, r.Path)
			continue
		}
		fmt.Fprintf(w, "  WARN skill skipped (%s): %s\n", r.SanitisedReason, r.Path)
	}
}

// SkipEnv builds a CODEX_SKIP_SKILLS-style comma list for paths the
// sanitiser could not rewrite. Empty list returns empty string so
// callers can `os.Setenv` unconditionally.
func SkipEnv(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return strings.Join(paths, ",")
}
