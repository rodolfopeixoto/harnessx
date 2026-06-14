// SPDX-License-Identifier: MIT

// Package plan renders the plan-confirmation template (spec §9).
package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

type Plan struct {
	ID                  string
	GeneratedAt         time.Time
	SpecID              string
	Mode                domain.Mode
	Summary             string
	DetectedStack       []string
	RelevantContext     []string
	ProposedApproach    string
	FilesLikelyToChange []string
	TestsToAddOrUpdate  []string
	SensorsToRun        []string
	SecurityChecks      []string
	Risks               []string
	Questions           []string
	EstimatedCostUSD    float64
	EstimatedTime       string
	ConfirmationStatus  string // "pending" | "approved" | "denied"
	AgentChain          []string
}

const tmplBody = `# Plan {{.ID}}

- **Spec:** {{.SpecID}}
- **Mode:** {{.Mode}}
- **Generated:** {{.GeneratedAt.Format "2006-01-02T15:04:05Z07:00"}}
- **Estimated cost:** ${{printf "%.2f" .EstimatedCostUSD}}
- **Estimated time:** {{.EstimatedTime}}
- **Confirmation:** {{.ConfirmationStatus}}

## 1. Summary
{{.Summary}}

## 2. Detected intent
{{.Mode}}

## 3. Detected stack
{{joinOrNone .DetectedStack ", "}}

## 4. Relevant project context
{{listOrNone .RelevantContext}}

## 5. Proposed approach
{{or .ProposedApproach "_TODO: describe the implementation approach._"}}

## 6. Files likely to change
{{listOrNone .FilesLikelyToChange}}

## 7. Tests to create or update
{{listOrNone .TestsToAddOrUpdate}}

## 8. Sensors to run
{{listOrNone .SensorsToRun}}

## 9. Security checks
{{listOrNone .SecurityChecks}}

## 10. Risks
{{listOrNone .Risks}}

## 11. Questions
{{listOrNone .Questions}}

## 12. Estimated cost
${{printf "%.2f" .EstimatedCostUSD}}

## 13. Estimated time
{{.EstimatedTime}}

## 14. Agent chain
{{joinOrNone .AgentChain " → "}}
`

func list(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for _, it := range items {
		b.WriteString("- ")
		b.WriteString(it)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func listOrNone(items []string) string {
	if len(items) == 0 {
		return "_(none identified)_"
	}
	return list(items)
}

func joinOrNone(items []string, sep string) string {
	if len(items) == 0 {
		return "_(none)_"
	}
	return strings.Join(items, sep)
}

var tmpl = template.Must(template.New("plan").Funcs(template.FuncMap{
	"list":       list,
	"listOrNone": listOrNone,
	"joinOrNone": joinOrNone,
}).Parse(tmplBody))

func New(specID string, mode domain.Mode) Plan {
	return Plan{
		ID: ids.New(), GeneratedAt: time.Now().UTC(),
		SpecID: specID, Mode: mode, ConfirmationStatus: "pending",
		EstimatedTime: "unknown",
	}
}

func (p Plan) Write(root string) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "plans")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(dir, p.ID+".md")
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := tmpl.Execute(f, p); err != nil {
		return "", err
	}
	return out, nil
}

func LatestPlanPath(root string) (string, error) {
	dir := filepath.Join(paths.HarnessDir(root), "artifacts", "plans")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var newest string
	var newestMod time.Time
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestMod) {
			newestMod = info.ModTime()
			newest = filepath.Join(dir, e.Name())
		}
	}
	if newest == "" {
		return "", fmt.Errorf("no plans under %s", dir)
	}
	return newest, nil
}
