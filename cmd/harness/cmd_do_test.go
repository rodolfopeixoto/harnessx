// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ropeixoto/harnessx/internal/router"
	"github.com/ropeixoto/harnessx/internal/taskgraph"
)

func TestHandoffBlock(t *testing.T) {
	steps := []plannedStep{
		{task: taskgraph.Task{Kind: taskgraph.KindScaffold}, chosen: "deterministic:scaffold:python"},
		{task: taskgraph.Task{Kind: taskgraph.KindCode}, chosen: "adapter:claude"},
	}
	results := []string{"scaffold-dry", "workflow-status:applied"}
	out := handoff("orig prompt", steps, results)
	for _, want := range []string{"Past steps", "orig prompt", "scaffold-dry", "workflow-status:applied", "This step"} {
		if !strings.Contains(out, want) {
			t.Errorf("handoff missing %q\n--- got ---\n%s", want, out)
		}
	}
}

func TestDeterministicMatch(t *testing.T) {
	cases := []struct {
		task taskgraph.Task
		want string
	}{
		{taskgraph.Task{Kind: taskgraph.KindScaffold, Lang: "python"}, "scaffold:python"},
		{taskgraph.Task{Kind: taskgraph.KindScaffold}, ""},
		{taskgraph.Task{Kind: taskgraph.KindLint}, "sensor:lint"},
		{taskgraph.Task{Kind: taskgraph.KindTest}, "sensor:test"},
		{taskgraph.Task{Kind: taskgraph.KindSecrets}, "sensor:secrets"},
		{taskgraph.Task{Kind: taskgraph.KindCode}, ""},
		{taskgraph.Task{Kind: taskgraph.KindImage}, ""},
	}
	for _, c := range cases {
		got := deterministicMatch(c.task)
		if got != c.want {
			t.Errorf("kind=%s lang=%q: want %q, got %q", c.task.Kind, c.task.Lang, c.want, got)
		}
	}
}

func TestTruncStr(t *testing.T) {
	if got := truncStr("hello world", 50); got != "hello world" {
		t.Errorf("short string should pass through: got %q", got)
	}
	if got := truncStr("hello world", 5); got != "hell…" {
		t.Errorf("long string trunc: got %q", got)
	}
}

func TestPrintPlanLowConfidenceWarning(t *testing.T) {
	steps := []plannedStep{
		{task: taskgraph.Task{Kind: taskgraph.KindGeneric, Confidence: 0.3}, chosen: "adapter:claude"},
	}
	var buf bytes.Buffer
	printPlan(&buf, steps)
	if !strings.Contains(buf.String(), "low classification confidence") {
		t.Error("expected low-confidence warning")
	}
}

func TestPrintPlanNoWarningOnHighConfidence(t *testing.T) {
	steps := []plannedStep{
		{task: taskgraph.Task{Kind: taskgraph.KindScaffold, Confidence: 1.0}, chosen: "deterministic:scaffold:python"},
	}
	var buf bytes.Buffer
	printPlan(&buf, steps)
	if strings.Contains(buf.String(), "low classification confidence") {
		t.Error("did not expect warning at confidence=1.0")
	}
}

func TestEmitJSONIncludesSchemaVersion(t *testing.T) {
	steps := []plannedStep{
		{task: taskgraph.Task{Kind: taskgraph.KindCode, Tags: []string{"code"}, Prompt: "x", Confidence: 0.7},
			choice: router.Choice{AdapterID: "claude"}, chosen: "adapter:claude"},
	}
	var buf bytes.Buffer
	if err := emitJSON(&buf, "x", steps); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "\"schema_version\": 1") {
		t.Errorf("missing schema_version: %s", got)
	}
	if !strings.Contains(got, "\"adapter_id\": \"claude\"") {
		t.Errorf("missing adapter_id: %s", got)
	}
}

func TestEmitDoJSONIncludesResultsAndSchema(t *testing.T) {
	steps := []plannedStep{
		{task: taskgraph.Task{Kind: taskgraph.KindScaffold, Prompt: "scaffold python", Lang: "python"},
			chosen: "deterministic:scaffold:python"},
	}
	var buf bytes.Buffer
	if err := emitDoJSON(&buf, "orig", steps, []string{"scaffold-dry"}, "/path/to/do.md"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"\"schema_version\": 1",
		"\"report_path\": \"/path/to/do.md\"",
		"\"results\":",
		"\"scaffold-dry\"",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}
