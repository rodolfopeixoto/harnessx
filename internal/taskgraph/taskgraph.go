// SPDX-License-Identifier: MIT

// Package taskgraph decomposes a free-form prompt into a sequence of
// typed tasks. Rule-based by default; the LLM decomposer is opt-in via
// Options.UseLLM and lives in a separate file so a build without the
// agents package still gets the deterministic decomposer.
package taskgraph

import (
	"regexp"
	"strings"
)

// Kind labels a task by the dominant operation it performs. Used by the
// router to pick a deterministic implementation (scaffold, sensor) or
// the best LLM adapter.
type Kind string

const (
	KindScaffold Kind = "scaffold" // deterministic: scaffoldpkg
	KindLint     Kind = "lint"     // deterministic: sensorcmd
	KindTest     Kind = "test"     // deterministic: sensorcmd
	KindFormat   Kind = "format"   // deterministic: gofmt / ruff / etc.
	KindSecrets  Kind = "secrets"  // deterministic: secrets_scan sensor
	KindCode     Kind = "code"     // LLM: write or modify code
	KindRefactor Kind = "refactor" // LLM: structural change
	KindDocs     Kind = "docs"     // LLM: write docs / README
	KindReview   Kind = "review"   // LLM: review diff / PR
	KindImage    Kind = "image"    // LLM (vision-capable): generate image
	KindVision   Kind = "vision"   // LLM (vision-capable): analyse image
	KindSearch   Kind = "search"   // LLM (search-capable): web/repo search
	KindData     Kind = "data"     // LLM: SQL / data manipulation
	KindShell    Kind = "shell"    // deterministic: passthrough
	KindGeneric  Kind = "generic"  // fallback when no rule matches
)

// Task is one node in the decomposed graph. Tags are the controlled
// vocabulary the router scores against adapter.Capabilities.Strengths.
type Task struct {
	Kind       Kind
	Tags       []string
	Prompt     string
	Lang       string // populated for KindScaffold
	DependsOn  []int  // indices into the decomposed slice
	Confidence float64
}

// Options tunes the decomposer.
type Options struct {
	UseLLM bool // future: route to a cheap LLM when rules return Generic
}

// Decompose splits prompt by clauses ("and", "then", ",", ";") and
// classifies each clause with a small rules set. Always returns at
// least one task.
func Decompose(prompt string, _ Options) []Task {
	clauses := splitClauses(prompt)
	tasks := make([]Task, 0, len(clauses))
	for _, c := range clauses {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		tasks = append(tasks, classify(c))
	}
	if len(tasks) == 0 {
		tasks = []Task{classify(prompt)}
	}
	return tasks
}

var clauseSplitter = regexp.MustCompile(`(?i)\s*(?:,| and then | then |, then |; |\sand\s)\s*`)

func splitClauses(s string) []string {
	parts := clauseSplitter.Split(s, -1)
	if len(parts) == 1 {
		return parts
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

type rule struct {
	pattern *regexp.Regexp
	kind    Kind
	tags    []string
	lang    string // when set, applies to scaffold detection
	score   float64
}

var rules = []rule{
	// Deterministic (no LLM): scaffold
	{regexp.MustCompile(`(?i)scaffold(?:s| a)? python|fastapi|django|flask`), KindScaffold, []string{"scaffold", "code"}, "python", 1.0},
	{regexp.MustCompile(`(?i)scaffold(?:s| a)? (?:go|golang)\b`), KindScaffold, []string{"scaffold", "code"}, "go", 1.0},
	{regexp.MustCompile(`(?i)scaffold(?:s| a)? ruby|sinatra|rails`), KindScaffold, []string{"scaffold", "code"}, "ruby", 1.0},
	{regexp.MustCompile(`(?i)scaffold(?:s| a)? rust|axum`), KindScaffold, []string{"scaffold", "code"}, "rust", 1.0},
	{regexp.MustCompile(`(?i)scaffold(?:s| a)? react|vite|next`), KindScaffold, []string{"scaffold", "code"}, "react", 1.0},
	{regexp.MustCompile(`(?i)\bscaffold\b`), KindScaffold, []string{"scaffold", "code"}, "", 0.6},

	// Deterministic: sensors
	{regexp.MustCompile(`(?i)\brun\b.*\blint(ers)?\b`), KindLint, []string{"lint", "code"}, "", 1.0},
	{regexp.MustCompile(`(?i)\b(only run|just run|run)\b.*\btests?\b`), KindTest, []string{"test", "code"}, "", 1.0},
	{regexp.MustCompile(`(?i)\bformat\b`), KindFormat, []string{"format", "code"}, "", 0.9},
	{regexp.MustCompile(`(?i)\b(scan )?(for )?secrets?\b`), KindSecrets, []string{"secrets"}, "", 0.9},

	// LLM-required tasks
	{regexp.MustCompile(`(?i)\b(image|banner|logo|illustration|picture)\b.*(generate|create|render|design)`), KindImage, []string{"image"}, "", 1.0},
	{regexp.MustCompile(`(?i)(generate|create|render|design|build|draw|make)\b.*\b(image|banner|logo|illustration|picture|hero)`), KindImage, []string{"image"}, "", 1.0},
	{regexp.MustCompile(`(?i)\b(this|the)\b.*\b(image|mockup|screenshot)\b`), KindVision, []string{"vision"}, "", 0.9},
	{regexp.MustCompile(`(?i)\brefactor\b`), KindRefactor, []string{"refactor", "code"}, "", 1.0},
	{regexp.MustCompile(`(?i)\b(review|audit)\b.*\b(diff|pr|pull request|code)\b`), KindReview, []string{"review", "code"}, "", 1.0},
	{regexp.MustCompile(`(?i)\b(write|update|add)\b.*\b(docs|documentation|readme)\b`), KindDocs, []string{"docs"}, "", 1.0},
	{regexp.MustCompile(`(?i)\bsearch\b.*\b(web|internet|docs)\b`), KindSearch, []string{"search"}, "", 1.0},
	{regexp.MustCompile(`(?i)\b(sql|query|migration|schema)\b`), KindData, []string{"data", "sql"}, "", 0.8},

	// Generic code change
	{regexp.MustCompile(`(?i)\b(add|implement|create|build|expose|introduce)\b`), KindCode, []string{"code"}, "", 0.7},
	{regexp.MustCompile(`(?i)\bfix\b`), KindCode, []string{"code"}, "", 0.7},
}

func classify(clause string) Task {
	for _, r := range rules {
		if r.pattern.MatchString(clause) {
			return Task{
				Kind:       r.kind,
				Tags:       r.tags,
				Prompt:     clause,
				Lang:       r.lang,
				Confidence: r.score,
			}
		}
	}
	return Task{Kind: KindGeneric, Tags: []string{"code"}, Prompt: clause, Confidence: 0.3}
}
