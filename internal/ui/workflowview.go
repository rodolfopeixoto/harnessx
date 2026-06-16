// SPDX-License-Identifier: MIT

package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Phase tags a workflow stage so the presenter can colour it
// consistently. Order mirrors the workflow pipeline.
type Phase string

const (
	PhaseIntent  Phase = "INTENT"
	PhasePlan    Phase = "PLAN"
	PhaseRoute   Phase = "ROUTE"
	PhaseAgent   Phase = "AGENT"
	PhaseSensors Phase = "SENSORS"
	PhaseBudget  Phase = "BUDGET"
	PhaseReport  Phase = "REPORT"
	PhaseLoop    Phase = "LOOP"
)

// Status of a phase end. ok = green ✓, warn = yellow ⚠, fail = red ✗,
// info = muted ·.
type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
	StatusInfo Status = "info"
)

// Presenter renders workflow output. richPresenter uses lipgloss
// colors + section headers; plainPresenter falls back to grep-friendly
// `[PHASE] ...` lines. Auto-pick is based on ui.IsPlain().
type Presenter interface {
	Start(phase Phase, headline string)
	Detail(line string)
	End(phase Phase, status Status, summary string)
	Note(line string)
}

func NewWorkflowView(out io.Writer) Presenter {
	if IsPlain() {
		return &plainPresenter{out: out}
	}
	return &richPresenter{out: out}
}

type plainPresenter struct {
	out io.Writer
}

func (p *plainPresenter) Start(phase Phase, headline string) {
	fmt.Fprintf(p.out, "[%s] %s\n", phase, headline)
}

func (p *plainPresenter) Detail(line string) {
	fmt.Fprintf(p.out, "        %s\n", line)
}

func (p *plainPresenter) End(phase Phase, status Status, summary string) {
	if summary == "" {
		return
	}
	fmt.Fprintf(p.out, "[%s] %s — %s\n", phase, string(status), summary)
}

func (p *plainPresenter) Note(line string) {
	fmt.Fprintf(p.out, "  %s\n", line)
}

type richPresenter struct {
	out io.Writer
}

var (
	phaseTag = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7DD3FC")).Width(9)
	headline = lipgloss.NewStyle().Bold(true)
	detail   = lipgloss.NewStyle().Faint(true)
	okIcon   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50")).Render("✓")
	warnIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300")).Render("⚠")
	failIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#E53935")).Render("✗")
	infoIcon = lipgloss.NewStyle().Faint(true).Render("·")
)

func (p *richPresenter) Start(phase Phase, headlineText string) {
	tag := phaseTag.Render(strings.ToLower(string(phase)))
	fmt.Fprintf(p.out, "%s %s %s\n", infoIcon, tag, headline.Render(headlineText))
}

func (p *richPresenter) Detail(line string) {
	fmt.Fprintf(p.out, "          %s\n", detail.Render(line))
}

func (p *richPresenter) End(phase Phase, status Status, summary string) {
	if summary == "" {
		return
	}
	icon := infoIcon
	switch status {
	case StatusOK:
		icon = okIcon
	case StatusWarn:
		icon = warnIcon
	case StatusFail:
		icon = failIcon
	}
	tag := phaseTag.Render(strings.ToLower(string(phase)))
	fmt.Fprintf(p.out, "%s %s %s\n", icon, tag, summary)
}

func (p *richPresenter) Note(line string) {
	fmt.Fprintf(p.out, "          %s\n", line)
}
