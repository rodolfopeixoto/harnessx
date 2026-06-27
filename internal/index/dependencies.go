// SPDX-License-Identifier: MIT

package index

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// BuildDependencies classifies dependencies per ecosystem. It does not
// attempt deep version resolution — that's the job of stack-native tools.
// We only need enough fidelity to feed routing/sensor decisions.
func BuildDependencies(root string, stacks []Stack) Dependencies {
	d := Dependencies{Ecosystems: map[string]Ecosystem{}}

	if e, ok := readNodeDeps(filepath.Join(root, "package.json")); ok {
		d.Ecosystems["node"] = e
	}
	if e, ok := readGoDeps(filepath.Join(root, "go.mod")); ok {
		d.Ecosystems["go"] = e
	}
	if e, ok := readGemfileLock(filepath.Join(root, "Gemfile.lock")); ok {
		d.Ecosystems["ruby"] = e
	} else if e, ok := readGemfile(filepath.Join(root, "Gemfile")); ok {
		d.Ecosystems["ruby"] = e
	}
	if e, ok := readRequirementsTxt(filepath.Join(root, "requirements.txt")); ok {
		d.Ecosystems["python"] = e
	} else if e, ok := readPyProjectToml(filepath.Join(root, "pyproject.toml")); ok {
		d.Ecosystems["python"] = e
	}
	if _, err := os.Stat(filepath.Join(root, "Cargo.toml")); err == nil {
		d.Ecosystems["rust"] = Ecosystem{Manifest: "Cargo.toml"}
	}
	_ = stacks // reserved for future per-stack hints
	return d
}

func readNodeDeps(path string) (Ecosystem, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Ecosystem{}, false
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		return Ecosystem{}, false
	}
	e := Ecosystem{Manifest: "package.json"}
	for _, k := range sortedKeysMap(pkg.Dependencies) {
		e.Runtime = append(e.Runtime, DependencyEntry{Name: k, Version: pkg.Dependencies[k]})
	}
	for _, k := range sortedKeysMap(pkg.DevDependencies) {
		e.Dev = append(e.Dev, DependencyEntry{Name: k, Version: pkg.DevDependencies[k]})
	}
	e.Count = len(e.Runtime) + len(e.Dev)
	return e, true
}

var goModRequireLine = regexp.MustCompile(`^\s*([^\s/]+/[^\s]+|[^\s]+)\s+(v[0-9][^\s]*)`)

func readGoDeps(path string) (Ecosystem, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Ecosystem{}, false
	}
	defer f.Close()
	e := Ecosystem{Manifest: "go.mod"}
	scanner := bufio.NewScanner(f)
	inBlock := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "require ("):
			inBlock = true
			continue
		case line == ")":
			inBlock = false
			continue
		case strings.HasPrefix(line, "require ") && !strings.HasSuffix(line, "("):
			line = strings.TrimPrefix(line, "require ")
		case !inBlock:
			continue
		}
		m := goModRequireLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		// Indirect deps live alongside direct deps; we capture both.
		entry := DependencyEntry{Name: m[1], Version: m[2]}
		if strings.Contains(line, "// indirect") {
			e.Dev = append(e.Dev, entry)
		} else {
			e.Runtime = append(e.Runtime, entry)
		}
	}
	sortDeps(e.Runtime)
	sortDeps(e.Dev)
	e.Count = len(e.Runtime) + len(e.Dev)
	return e, true
}

var gemLine = regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)

func readGemfile(path string) (Ecosystem, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Ecosystem{}, false
	}
	defer f.Close()
	e := Ecosystem{Manifest: "Gemfile"}
	scanner := bufio.NewScanner(f)
	groupDev := false
	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "group ") && (strings.Contains(trim, ":development") || strings.Contains(trim, ":test")):
			groupDev = true
		case strings.HasPrefix(trim, "end"):
			groupDev = false
		}
		m := gemLine.FindStringSubmatch(trim)
		if m == nil {
			continue
		}
		entry := DependencyEntry{Name: m[1], Version: m[2]}
		if groupDev {
			e.Dev = append(e.Dev, entry)
		} else {
			e.Runtime = append(e.Runtime, entry)
		}
	}
	sortDeps(e.Runtime)
	sortDeps(e.Dev)
	e.Count = len(e.Runtime) + len(e.Dev)
	return e, true
}

var gemlockSpecLine = regexp.MustCompile(`^\s{4}([a-zA-Z0-9_\-]+)\s+\(([^)]+)\)`)

func readGemfileLock(path string) (Ecosystem, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Ecosystem{}, false
	}
	defer f.Close()
	e := Ecosystem{Manifest: "Gemfile.lock"}
	scanner := bufio.NewScanner(f)
	inSpecs := false
	seen := map[string]bool{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "specs:" {
			inSpecs = true
			continue
		}
		if inSpecs && len(line) > 0 && line[0] != ' ' {
			inSpecs = false
		}
		if !inSpecs {
			continue
		}
		m := gemlockSpecLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		e.Runtime = append(e.Runtime, DependencyEntry{Name: m[1], Version: m[2]})
	}
	sortDeps(e.Runtime)
	e.Count = len(e.Runtime)
	return e, true
}

var reqLine = regexp.MustCompile(`^([A-Za-z0-9_\-.]+)\s*(?:[=<>!~]=?\s*([^\s;]+))?`)

func readRequirementsTxt(path string) (Ecosystem, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Ecosystem{}, false
	}
	defer f.Close()
	e := Ecosystem{Manifest: "requirements.txt"}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		m := reqLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		e.Runtime = append(e.Runtime, DependencyEntry{Name: m[1], Version: m[2]})
	}
	sortDeps(e.Runtime)
	e.Count = len(e.Runtime)
	return e, true
}

var pyProjectDepLine = regexp.MustCompile(`^\s*"([A-Za-z0-9_\-.]+)\s*(?:[=<>!~][^"]*)?"\s*,?\s*$`)

func readPyProjectToml(path string) (Ecosystem, bool) {
	f, err := os.Open(path)
	if err != nil {
		return Ecosystem{}, false
	}
	defer f.Close()
	e := Ecosystem{Manifest: "pyproject.toml"}
	scanner := bufio.NewScanner(f)
	section := ""
	for scanner.Scan() {
		raw := strings.TrimRight(scanner.Text(), "\r\n")
		trim := strings.TrimSpace(raw)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(trim, "[") && strings.HasSuffix(trim, "]") {
			section = strings.Trim(trim, "[]")
			continue
		}
		if section == "project" && strings.HasPrefix(trim, "dependencies") {
			collectInlineList(scanner, trim, &e.Runtime)
			continue
		}
		if section == "project.optional-dependencies" || strings.HasPrefix(section, "project.optional-dependencies") {
			collectInlineList(scanner, trim, &e.Dev)
			continue
		}
		if section == "tool.poetry.dependencies" {
			if name := poetryDepName(trim); name != "" {
				e.Runtime = append(e.Runtime, DependencyEntry{Name: name})
			}
			continue
		}
		if section == "tool.poetry.dev-dependencies" || section == "tool.poetry.group.dev.dependencies" {
			if name := poetryDepName(trim); name != "" {
				e.Dev = append(e.Dev, DependencyEntry{Name: name})
			}
			continue
		}
	}
	sortDeps(e.Runtime)
	sortDeps(e.Dev)
	e.Count = len(e.Runtime) + len(e.Dev)
	if e.Count == 0 {
		return e, true
	}
	return e, true
}

func collectInlineList(scanner *bufio.Scanner, first string, dst *[]DependencyEntry) {
	if idx := strings.Index(first, "["); idx >= 0 {
		first = first[idx+1:]
	}
	for {
		body := strings.TrimSpace(first)
		closed := strings.HasSuffix(body, "]")
		if closed {
			body = strings.TrimSpace(strings.TrimSuffix(body, "]"))
		}
		body = strings.TrimSuffix(strings.TrimSpace(body), ",")
		for _, part := range strings.Split(body, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if m := pyProjectDepLine.FindStringSubmatch(part); m != nil {
				*dst = append(*dst, DependencyEntry{Name: m[1]})
			} else if name := pyProjectFreeDep(part); name != "" {
				*dst = append(*dst, DependencyEntry{Name: name})
			}
		}
		if closed {
			return
		}
		if !scanner.Scan() {
			return
		}
		first = scanner.Text()
	}
}

var pyProjectFreeRE = regexp.MustCompile(`"([A-Za-z0-9_\-.]+)`)

func pyProjectFreeDep(part string) string {
	m := pyProjectFreeRE.FindStringSubmatch(part)
	if m == nil {
		return ""
	}
	return m[1]
}

var poetryNameRE = regexp.MustCompile(`^([A-Za-z0-9_\-.]+)\s*=`)

func poetryDepName(line string) string {
	m := poetryNameRE.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	if m[1] == "python" {
		return ""
	}
	return m[1]
}

func sortDeps(d []DependencyEntry) {
	sort.Slice(d, func(i, j int) bool { return d[i].Name < d[j].Name })
}

func sortedKeysMap(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
