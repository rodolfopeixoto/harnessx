// SPDX-License-Identifier: MIT

package auditrun

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type ReportInput struct {
	Timestamp string
	BaseURL   string
	Summary   Summary
	Features  []Feature
	Results   []Result
	Backlog   []BacklogItem
}

func WriteHTMLReport(path string, in ReportInput) error {
	tpl, err := template.New("audit").Funcs(template.FuncMap{
		"pct":        func(f float64) string { return fmt.Sprintf("%.1f%%", f*100) },
		"join":       func(values []string) string { return strings.Join(values, ", ") },
		"hasResults": func(s string) bool { return s != "" },
	}).Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return tpl.Execute(fh, in)
}

func WriteBacklog(path string, items []BacklogItem) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Fix backlog\n\n")
	if len(items) == 0 {
		b.WriteString("no failing features.\n")
		return os.WriteFile(path, []byte(b.String()), 0o644)
	}
	for _, it := range items {
		fmt.Fprintf(&b, "## %s — %s\n\n", it.Severity, it.Feature)
		fmt.Fprintf(&b, "- id: %s\n- route: %s\n- role: %s\n- viewport: %s\n- failure: %s\n", it.ID, it.Route, it.Role, it.Viewport, it.FailureType)
		if it.Reproduce != "" {
			fmt.Fprintf(&b, "- reproduce: `%s`\n", it.Reproduce)
		}
		if it.Expected != "" {
			fmt.Fprintf(&b, "- expected: %s\n", it.Expected)
		}
		if it.Actual != "" {
			fmt.Fprintf(&b, "- actual: %s\n", it.Actual)
		}
		if it.Screenshot != "" {
			fmt.Fprintf(&b, "- screenshot: `%s`\n", it.Screenshot)
		}
		if it.DiffImage != "" {
			fmt.Fprintf(&b, "- diff: `%s`\n", it.DiffImage)
		}
		if it.Suggestion != "" {
			fmt.Fprintf(&b, "- suggestion: %s\n", it.Suggestion)
		}
		if it.AcceptCriteria != "" {
			fmt.Fprintf(&b, "- accept criteria: %s\n", it.AcceptCriteria)
		}
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func TerminalSummary(in ReportInput, artifacts map[string]string) string {
	c := in.Summary.Counts
	v := in.Summary.Visual
	var b strings.Builder
	fmt.Fprintln(&b, "Audit finished")
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "Functional:")
	for _, key := range []string{
		constants.AuditStatusPassed,
		constants.AuditStatusFailed,
		constants.AuditStatusPartial,
		constants.AuditStatusBlocked,
		constants.AuditStatusNotImplemented,
	} {
		fmt.Fprintf(&b, "%s: %d\n", key, c[key])
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "Visual:")
	for _, key := range []string{
		constants.AuditVisualPassed,
		constants.AuditVisualMinorDiff,
		constants.AuditVisualMajorDiff,
		constants.AuditStatusVisualBroken,
		constants.AuditStatusLayoutCollapsed,
		constants.AuditStatusWrongScreen,
	} {
		fmt.Fprintf(&b, "%s: %d\n", key, v[key])
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "Technical:")
	for _, key := range []string{
		constants.AuditStatusAPIError,
		constants.AuditStatusConsoleError,
		constants.AuditStatusPermissionError,
		constants.AuditStatusSelectorMissing,
		constants.AuditStatusDataMissing,
	} {
		fmt.Fprintf(&b, "%s: %d\n", key, c[key])
	}
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "Artifacts:")
	for _, key := range []string{"PDF", "HTML", "Screenshots current", "Screenshots reference", "Diffs", "JSON", "Backlog", "Log"} {
		if v := artifacts[key]; v != "" {
			fmt.Fprintf(&b, "%s: %s\n", key, v)
		}
	}
	return b.String()
}

const htmlTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>HarnessX stack audit — {{.Timestamp}}</title>
<style>
  body { font-family: -apple-system, system-ui, sans-serif; margin: 0; color: #0F172A; }
  header { background: #4338CA; color: white; padding: 24px; }
  main { padding: 24px; max-width: 1080px; margin: 0 auto; }
  h1, h2, h3 { margin: 16px 0 8px; }
  table { border-collapse: collapse; width: 100%; margin: 12px 0; }
  th, td { border: 1px solid #E4E4E7; padding: 8px; text-align: left; font-size: 13px; }
  th { background: #F1F5F9; }
  .pill { display: inline-block; padding: 2px 8px; border-radius: 999px; font-size: 11px; font-weight: 600; }
  .passed { background: #DCFCE7; color: #166534; }
  .failed { background: #FEE2E2; color: #991B1B; }
  .partial { background: #FEF9C3; color: #854D0E; }
  .blocked { background: #E0E7FF; color: #3730A3; }
  .other { background: #F1F5F9; color: #475569; }
  img { max-width: 100%; height: auto; border: 1px solid #E4E4E7; border-radius: 8px; margin: 8px 0; }
</style>
</head>
<body>
<header>
  <h1>HarnessX stack audit</h1>
  <p>Run timestamp: {{.Timestamp}} · base URL: {{.BaseURL}}</p>
</header>
<main>
  <h2>Summary</h2>
  <table>
    <tr><th>Total features</th><td>{{.Summary.TotalFeatures}}</td></tr>
    <tr><th>Total results</th><td>{{.Summary.TotalResults}}</td></tr>
    <tr><th>Pass rate</th><td>{{pct .Summary.PassRate}}</td></tr>
  </table>
  <h2>Functional counts</h2>
  <table>
    <thead><tr><th>Status</th><th>Count</th></tr></thead>
    <tbody>
      {{range $k, $v := .Summary.Counts}}<tr><td>{{$k}}</td><td>{{$v}}</td></tr>{{end}}
    </tbody>
  </table>
  <h2>Visual counts</h2>
  <table>
    <thead><tr><th>Visual status</th><th>Count</th></tr></thead>
    <tbody>
      {{range $k, $v := .Summary.Visual}}<tr><td>{{$k}}</td><td>{{$v}}</td></tr>{{end}}
    </tbody>
  </table>
  <h2>Feature matrix</h2>
  <table>
    <thead><tr><th>ID</th><th>Feature</th><th>Route</th><th>Role</th><th>Priority</th><th>Viewports</th></tr></thead>
    <tbody>
      {{range .Features}}
      <tr>
        <td>{{.ID}}</td><td>{{.Name}}</td><td>{{.Route}}</td>
        <td>{{.Role}}</td><td>{{.Priority}}</td><td>{{join .Viewports}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
  <h2>Results</h2>
  <table>
    <thead><tr><th>Feature</th><th>Viewport</th><th>Status</th><th>HTTP</th><th>Reason</th><th>Screenshot</th></tr></thead>
    <tbody>
    {{range .Results}}
      <tr>
        <td>{{.FeatureID}}</td>
        <td>{{.Viewport}}</td>
        <td><span class="pill {{if eq .Status "passed"}}passed{{else if eq .Status "failed"}}failed{{else if eq .Status "partial"}}partial{{else if eq .Status "blocked"}}blocked{{else}}other{{end}}">{{.Status}}</span></td>
        <td>{{.HTTPStatus}}</td>
        <td>{{.Reason}}</td>
        <td>{{if hasResults .Screenshot}}<img src="../{{.Screenshot}}" alt="{{.FeatureID}}-{{.Viewport}}">{{end}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  <h2>Backlog</h2>
  {{if .Backlog}}
  <table>
    <thead><tr><th>Severity</th><th>Feature</th><th>Failure</th><th>Reproduce</th></tr></thead>
    <tbody>
    {{range .Backlog}}
      <tr>
        <td><span class="pill {{if eq .Severity "P0"}}failed{{else if eq .Severity "P1"}}partial{{else}}other{{end}}">{{.Severity}}</span></td>
        <td>{{.Feature}} ({{.Viewport}})</td>
        <td>{{.FailureType}} — {{.Actual}}</td>
        <td><code>{{.Reproduce}}</code></td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{else}}<p>No failing features.</p>{{end}}
</main>
</body>
</html>`
