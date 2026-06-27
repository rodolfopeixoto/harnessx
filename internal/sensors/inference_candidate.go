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

type InferenceCandidateSensor struct {
	IDValue          string
	MinLinesForFlag  int
	MaxFuncLines     int
	FlagRegexComplex bool
	FlagBareTODO     bool
}

func (s InferenceCandidateSensor) ID() string                     { return s.IDValue }
func (s InferenceCandidateSensor) Category() Category             { return CatOther }
func (s InferenceCandidateSensor) Kind() Kind                     { return KindInferential }
func (s InferenceCandidateSensor) AppliesTo(_ index.Profile) bool { return true }

func (s InferenceCandidateSensor) Run(rc RunCtx) Result {
	start := time.Now()
	res := Result{ID: s.IDValue, Category: s.Category(), Kind: s.Kind()}
	if s.MinLinesForFlag == 0 {
		s.MinLinesForFlag = 200
	}
	walkCtx, cancel := context.WithTimeout(rc.Ctx, 60*time.Second)
	defer cancel()
	var findings []SmellFinding
	exts := map[string]bool{}
	for _, e := range []string{".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".rb", ".rs", ".java", ".kt", ".swift", ".ex", ".exs", ".php", ".cs", ".dart"} {
		exts[e] = true
	}
	_ = filepath.WalkDir(rc.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if walkCtx.Err() != nil {
			return walkCtx.Err()
		}
		if d.IsDir() {
			if isSkippedDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		findings = append(findings, s.scan(path, rc.Root)...)
		return nil
	})
	res.Duration = time.Since(start)
	res.Status = StatusPassed
	if len(findings) > 0 {
		res.Detail = fmt.Sprintf("%d candidate(s) for LLM review", len(findings))
		if rc.OutputDir != "" {
			_ = os.MkdirAll(rc.OutputDir, 0o755)
			p := filepath.Join(rc.OutputDir, s.IDValue+".log")
			_ = os.WriteFile(p, []byte(renderFindings(findings)), 0o644)
			res.OutputPath = p
		}
	} else {
		res.Detail = "no inference candidates"
	}
	return res
}

func (s InferenceCandidateSensor) scan(path, root string) []SmellFinding {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	rel, _ := filepath.Rel(root, path)
	var findings []SmellFinding
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	line := 0
	maxFunc := 0
	currentFunc := 0
	inFunc := false
	complexRegex := 0
	bareTODO := 0
	for scanner.Scan() {
		line++
		raw := scanner.Text()
		trim := strings.TrimSpace(raw)
		if isFuncDecl(trim) {
			if currentFunc > maxFunc {
				maxFunc = currentFunc
			}
			currentFunc = 0
			inFunc = true
		}
		if inFunc {
			currentFunc++
		}
		if s.FlagRegexComplex && containsComplexRegex(raw) {
			complexRegex++
		}
		if s.FlagBareTODO && strings.Contains(trim, "TODO") && !strings.Contains(trim, "(") && !strings.Contains(trim, "#") {
			bareTODO++
		}
	}
	if currentFunc > maxFunc {
		maxFunc = currentFunc
	}
	hasTest := strings.HasSuffix(path, "_test.go") || strings.Contains(rel, "tests/") || strings.Contains(rel, "test/") || strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".test.tsx") || strings.HasSuffix(path, ".spec.ts")
	if !hasTest && line >= s.MinLinesForFlag {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "untested_large_module",
			Msg: fmt.Sprintf("%d-line module with no co-located test file — candidate for LLM-assisted test generation", line),
		})
	}
	if !hasTest && s.MaxFuncLines > 0 && maxFunc > s.MaxFuncLines {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "untested_long_func",
			Msg: fmt.Sprintf("longest function is %d lines and untested — candidate for LLM review", maxFunc),
		})
	}
	if complexRegex > 0 {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "complex_regex",
			Msg: fmt.Sprintf("%d regex(es) with > 60 chars — candidate for LLM-assisted explanation/test", complexRegex),
		})
	}
	if bareTODO > 0 {
		findings = append(findings, SmellFinding{
			Path: rel, Line: 1, Kind: "bare_todo_cluster",
			Msg: fmt.Sprintf("%d bare TODOs without references — candidate for triage", bareTODO),
		})
	}
	return findings
}

func containsComplexRegex(line string) bool {
	lower := strings.ToLower(line)
	idx := strings.Index(lower, "regex")
	if idx < 0 {
		idx = strings.Index(lower, "compile(")
	}
	if idx < 0 {
		return false
	}
	rest := lower[idx:]
	if len(rest) > 200 {
		rest = rest[:200]
	}
	if strings.Count(rest, "\\") >= 3 || (strings.Count(rest, "[") >= 2 && strings.Count(rest, "(") >= 2) {
		return len(rest) > 60
	}
	return false
}
