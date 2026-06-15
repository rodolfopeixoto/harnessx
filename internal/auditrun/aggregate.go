// SPDX-License-Identifier: MIT

package auditrun

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

func Aggregate(results []Result, features []Feature) Summary {
	counts := map[string]int{}
	visual := map[string]int{}
	severity := map[string]int{}
	for _, r := range results {
		counts[r.Status]++
		if r.Visual != nil {
			visual[r.Visual.VisualStatus]++
		}
	}
	passed := counts[constants.AuditStatusPassed]
	total := len(results)
	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total)
	}
	for _, f := range features {
		severity[f.Priority]++
	}
	return Summary{
		GeneratedAt:   time.Now().UTC(),
		Counts:        counts,
		Visual:        visual,
		Severity:      severity,
		PassRate:      passRate,
		TotalFeatures: len(features),
		TotalResults:  total,
	}
}

func WriteJSONFile(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func ReadResults(path string) (Results, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Results{}, err
	}
	var r Results
	if err := json.Unmarshal(b, &r); err != nil {
		return Results{}, err
	}
	return r, nil
}

func BuildBacklog(results []Result, features []Feature) []BacklogItem {
	byID := map[string]Feature{}
	for _, f := range features {
		byID[f.ID] = f
	}
	var out []BacklogItem
	for _, r := range results {
		if r.Status == constants.AuditStatusPassed {
			continue
		}
		f := byID[r.FeatureID]
		sev := classifySeverity(f, r)
		out = append(out, BacklogItem{
			ID:             fmt.Sprintf("%s-%s", r.FeatureID, r.Viewport),
			Severity:       sev,
			Feature:        f.Name,
			Route:          f.Route,
			Role:           string(f.Role),
			Viewport:       r.Viewport,
			FailureType:    r.Status,
			Reproduce:      fmt.Sprintf("AUDIT_FEATURE=%s AUDIT_ROLE=%s bin/stack audit", f.ID, f.Role),
			Expected:       fmt.Sprintf("status=passed, http=%d", f.ExpectedHTTPStatus),
			Actual:         r.Reason,
			Screenshot:     r.Screenshot,
			DiffImage:      diffImageOf(r),
			Suggestion:     suggestionFor(r),
			AcceptCriteria: "audit re-run reports status=passed and visual_diff_pct <= 5",
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return severityRank(out[i].Severity) < severityRank(out[j].Severity)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func severityRank(s string) int {
	switch s {
	case constants.AuditSeverityP0:
		return 0
	case constants.AuditSeverityP1:
		return 1
	case constants.AuditSeverityP2:
		return 2
	case constants.AuditSeverityP3:
		return 3
	}
	return 4
}

func classifySeverity(f Feature, r Result) string {
	switch r.Status {
	case constants.AuditStatusPermissionError,
		constants.AuditStatusWrongScreen,
		constants.AuditStatusLayoutCollapsed,
		constants.AuditStatusVisualBroken:
		return constants.AuditSeverityP0
	case constants.AuditStatusAPIError, constants.AuditStatusConsoleError, constants.AuditStatusFailed:
		return constants.AuditSeverityP1
	case constants.AuditStatusPartial, constants.AuditStatusSelectorMissing, constants.AuditStatusDataMissing:
		return constants.AuditSeverityP2
	}
	if f.Priority != "" {
		return f.Priority
	}
	return constants.AuditSeverityP3
}

func diffImageOf(r Result) string {
	if r.Visual == nil {
		return ""
	}
	return r.Visual.DiffImage
}

func suggestionFor(r Result) string {
	switch r.Status {
	case constants.AuditStatusSelectorMissing:
		return "add a stable data-testid attribute to the page header or main action"
	case constants.AuditStatusAPIError:
		return "check the API handler in internal/adapters/http for the route in question"
	case constants.AuditStatusPermissionError:
		return "verify the role gate in web/dashboard/src/auth + backend middleware"
	case constants.AuditStatusVisualBroken, constants.AuditStatusLayoutCollapsed:
		return "inspect web/dashboard/src/ds tokens + the page CSS that wraps the failing component"
	case constants.AuditStatusDataMissing:
		return "seed fixture data via harness init + harness project add before re-running"
	}
	return "re-run the audit with AUDIT_HEADED=1 to inspect interactively"
}
