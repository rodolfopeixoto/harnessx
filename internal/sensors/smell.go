package sensors

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/index"
)

type SmellSensor struct {
	IDValue       string
	StacksV       []string
	MaxFileLines  int
	MaxFuncLines  int
	MaxNestDepth  int
	MaxMethodsCls int
	Extensions    []string
}

func (s SmellSensor) ID() string         { return s.IDValue }
func (s SmellSensor) Category() Category { return CatLint }
func (s SmellSensor) Kind() Kind         { return KindComputational }

func (s SmellSensor) AppliesTo(p index.Profile) bool {
	if len(s.StacksV) == 0 {
		return true
	}
	have := map[string]bool{}
	for _, st := range p.Stacks {
		have[st.Name] = true
	}
	for _, want := range s.StacksV {
		if have[want] {
			return true
		}
	}
	return false
}

type SmellFinding struct {
	Path string
	Line int
	Kind string
	Msg  string
}

func (s SmellSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.IDValue, Category: s.Category(), Kind: s.Kind()}
	var findings []SmellFinding
	walkCtx, cancel := context.WithTimeout(rc.Ctx, 60*time.Second)
	defer cancel()
	_ = filepath.WalkDir(rc.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if walkCtx.Err() != nil {
			return walkCtx.Err()
		}
		if d.IsDir() {
			name := d.Name()
			if isSkippedDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if !s.matchesExtension(path) {
			return nil
		}
		findings = append(findings, s.scanFile(path, rc.Root)...)
		return nil
	})
	res.Duration = time.Since(start)
	if len(findings) == 0 {
		res.Status = StatusPassed
		res.Detail = "no smells"
		return res
	}
	res.Status = StatusFailed
	res.Detail = fmt.Sprintf("%d smell(s)", len(findings))
	if rc.OutputDir != "" {
		_ = os.MkdirAll(rc.OutputDir, 0o755)
		p := filepath.Join(rc.OutputDir, s.IDValue+".log")
		_ = os.WriteFile(p, []byte(renderFindings(findings)), 0o644)
		res.OutputPath = p
	}
	return res
}

func isSkippedDir(name string) bool {
	switch name {
	case ".git", ".harness", "node_modules", "vendor", ".venv", "venv",
		"target", "dist", "build", "_build", "deps", ".gradle", ".idea",
		"bin", "obj":
		return true
	}
	return false
}

func (s SmellSensor) matchesExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, want := range s.Extensions {
		if ext == want {
			return true
		}
	}
	return false
}

func (s SmellSensor) scanFile(path, root string) []SmellFinding {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	rel, _ := filepath.Rel(root, path)
	var findings []SmellFinding
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	lineNo := 0
	methodCount := 0
	currentFuncStart := 0
	currentFuncDepth := 0
	maxNest := 0
	depth := 0
	inFunc := false
	funcStartLine := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")
		if isFuncDecl(trim) {
			methodCount++
			if !inFunc {
				inFunc = true
				funcStartLine = lineNo
				currentFuncStart = lineNo
				currentFuncDepth = depth
			}
		}
		depth += opens - closes
		if depth < 0 {
			depth = 0
		}
		if depth > maxNest {
			maxNest = depth
		}
		if inFunc && depth <= currentFuncDepth && lineNo > funcStartLine {
			funcLen := lineNo - currentFuncStart
			if s.MaxFuncLines > 0 && funcLen > s.MaxFuncLines {
				findings = append(findings, SmellFinding{
					Path: rel, Line: currentFuncStart, Kind: "long_method",
					Msg: fmt.Sprintf("function spans %d lines (limit %d)", funcLen, s.MaxFuncLines),
				})
			}
			inFunc = false
		}
	}
	if s.MaxFileLines > 0 && lineNo > s.MaxFileLines {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "long_file",
			Msg: fmt.Sprintf("file has %d lines (limit %d)", lineNo, s.MaxFileLines),
		})
	}
	if s.MaxNestDepth > 0 && maxNest > s.MaxNestDepth {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "deep_nesting",
			Msg: fmt.Sprintf("nesting depth %d exceeds limit %d", maxNest, s.MaxNestDepth),
		})
	}
	if s.MaxMethodsCls > 0 && methodCount > s.MaxMethodsCls {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "god_class",
			Msg: fmt.Sprintf("file declares %d methods/funcs (limit %d)", methodCount, s.MaxMethodsCls),
		})
	}
	return findings
}

func isFuncDecl(line string) bool {
	prefixes := []string{
		"func ", "fn ", "def ",
		"public ", "private ", "protected ", "internal ", "static ",
		"fun ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) && strings.Contains(line, "(") {
			return true
		}
	}
	return false
}

func renderFindings(items []SmellFinding) string {
	var b strings.Builder
	for _, f := range items {
		fmt.Fprintf(&b, "%s:%d [%s] %s\n", f.Path, f.Line, f.Kind, f.Msg)
	}
	return b.String()
}

func smellSensorsFor(stack string) []Sensor {
	exts := defaultExtensions(stack)
	if len(exts) == 0 {
		return nil
	}
	return []Sensor{
		SmellSensor{
			IDValue:       stack + "_smell",
			StacksV:       []string{stack},
			MaxFileLines:  600,
			MaxFuncLines:  60,
			MaxNestDepth:  6,
			MaxMethodsCls: 25,
			Extensions:    exts,
		},
	}
}

func defaultExtensions(stack string) []string {
	switch stack {
	case "go":
		return []string{".go"}
	case "python":
		return []string{".py"}
	case "node", "react", "nextjs", "vite":
		return []string{".ts", ".tsx", ".js", ".jsx"}
	case "ruby", "rails":
		return []string{".rb"}
	case "rust":
		return []string{".rs"}
	case "java":
		return []string{".java"}
	case "kotlin":
		return []string{".kt", ".kts"}
	case "swift":
		return []string{".swift"}
	case "elixir":
		return []string{".ex", ".exs"}
	case "php", "laravel", "symfony":
		return []string{".php"}
	case "dotnet":
		return []string{".cs", ".fs"}
	case "dart", "flutter":
		return []string{".dart"}
	}
	return nil
}
