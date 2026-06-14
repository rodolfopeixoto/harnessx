// SPDX-License-Identifier: MIT

package optimize

import "fmt"

// Compare returns the structured diff between two snapshots. From is the
// older snapshot; To is the newer one. Status semantics: numeric metrics
// improve when they decrease (lower disk use, fewer noisy logs, etc.).
func Compare(from, to Snapshot) Diff {
	d := Diff{From: from, To: to}
	rows := []DiffRow{
		intRow("deps_total", from.Deps.Total, to.Deps.Total),
		intRow("removal_candidates", len(from.Deps.Candidates), len(to.Deps.Candidates)),
		intRow("noisy_log_call_sites", from.Logs.TotalCallSites, to.Logs.TotalCallSites),
		int64Row("harness_dir_bytes", from.Disk.HarnessBytes, to.Disk.HarnessBytes),
		int64Row("project_bytes", from.Disk.ProjectBytes, to.Disk.ProjectBytes),
		int64Row("jsonl_log_bytes", from.Logs.JSONLBytes, to.Logs.JSONLBytes),
	}
	if from.Dockerfile != nil && to.Dockerfile != nil {
		rows = append(rows,
			intRow("dockerfile_run_steps", from.Dockerfile.RunSteps, to.Dockerfile.RunSteps),
			intRow("dockerfile_findings", len(from.Dockerfile.Findings), len(to.Dockerfile.Findings)),
		)
	}
	d.Rows = rows
	return d
}

func intRow(name string, a, b int) DiffRow {
	delta := b - a
	return DiffRow{
		Metric: name, Before: a, After: b,
		Delta: formatSignedInt(delta), Status: statusFor(delta),
	}
}

func int64Row(name string, a, b int64) DiffRow {
	delta := b - a
	return DiffRow{
		Metric: name, Before: a, After: b,
		Delta: formatSignedInt64(delta), Status: statusFor64(delta),
	}
}

func statusFor(delta int) string {
	switch {
	case delta < 0:
		return "improved"
	case delta > 0:
		return "regressed"
	default:
		return "unchanged"
	}
}

func statusFor64(delta int64) string {
	switch {
	case delta < 0:
		return "improved"
	case delta > 0:
		return "regressed"
	default:
		return "unchanged"
	}
}

func formatSignedInt(n int) string {
	if n > 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}

func formatSignedInt64(n int64) string {
	if n > 0 {
		return fmt.Sprintf("+%d", n)
	}
	return fmt.Sprintf("%d", n)
}
