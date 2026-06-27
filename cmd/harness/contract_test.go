package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type contractRule struct {
	command        []string
	mustContain    []string
	mustNotContain []string
	mustNotClaim   []string
	exitOK         bool
	stdoutNotEmpty bool
}

func TestContractsHelp(t *testing.T) {
	cases := []contractRule{
		{
			command:        []string{"ask", "--help"},
			mustContain:    []string{"evidence", "deterministic"},
			mustNotClaim:   []string{"answer the question", "synthesizes an answer", "generates an answer"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"plan", "--help"},
			mustContain:    []string{"deterministic"},
			mustNotClaim:   []string{"AI-generated plan", "LLM-generated plan", "model-generated"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"health", "show", "--help"},
			mustContain:    []string{"placeholder"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"stack", "status", "--help"},
			mustContain:    []string{"exits 0"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"cost-compare", "--help"},
			mustContain:    []string{"No agent is invoked", "0 LLM tokens"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"chat", "--help"},
			mustContain:    []string{"--once", "--route"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"version"},
			mustContain:    []string{"v"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"config", "sources", "--help"},
			mustContain:    []string{"update_repo"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
		{
			command:        []string{"install", "tools", "--help"},
			mustContain:    []string{"swift", "kotlin", "elixir"},
			exitOK:         true,
			stdoutNotEmpty: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(strings.Join(c.command, " "), func(t *testing.T) {
			runContract(t, c)
		})
	}
}

func TestUnknownCommandExitsNonZero(t *testing.T) {
	root := newRoot()
	root.SetArgs([]string{"this-is-not-a-real-command"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	if err := root.Execute(); err == nil {
		t.Fatalf("unknown command must exit non-zero, got nil err. stdout=%q", buf.String())
	}
}

func TestNaturalLanguagePositionalDoesNotFallthrough(t *testing.T) {
	root := newRoot()
	root.SetArgs([]string{"add a new endpoint please"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	err := root.Execute()
	if err == nil {
		t.Fatalf("natural-language positional must error, got nil. stdout=%q", buf.String())
	}
	combined := strings.ToLower(buf.String() + err.Error())
	if strings.Contains(combined, "detected intent: feature") {
		t.Fatalf("must not silently dispatch feature mode: %s", combined)
	}
}

func runContract(t *testing.T, c contractRule) {
	t.Helper()
	root := newRoot()
	root.SetArgs(c.command)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	err := root.Execute()
	combined := stdout.String() + stderr.String()
	if c.exitOK && err != nil {
		t.Fatalf("expected exit 0 for %v, got err=%v\noutput:\n%s", c.command, err, combined)
	}
	if c.stdoutNotEmpty && strings.TrimSpace(combined) == "" {
		t.Fatalf("expected non-empty output for %v", c.command)
	}
	for _, want := range c.mustContain {
		if !strings.Contains(combined, want) {
			t.Errorf("%v: output should mention %q\ngot:\n%s", c.command, want, combined)
		}
	}
	for _, banned := range c.mustNotContain {
		if strings.Contains(combined, banned) {
			t.Errorf("%v: output must NOT contain %q\ngot:\n%s", c.command, banned, combined)
		}
	}
	for _, lie := range c.mustNotClaim {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(lie))
		if re.MatchString(combined) {
			t.Errorf("%v: help text claims %q but the command is deterministic — this is the help-vs-behavior lie the audit caught\ngot:\n%s", c.command, lie, combined)
		}
	}
	_ = cobra.Command{}
}
