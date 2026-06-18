// SPDX-License-Identifier: MIT

package ui

import "github.com/charmbracelet/lipgloss"

var (
	Heading = lipgloss.NewStyle().Bold(true)
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50"))
	Warn    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300"))
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E53935"))
	Muted   = lipgloss.NewStyle().Faint(true)
	Info    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4"))
	Accent  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C4DFF")).Bold(true)
)

func MarkSuccess() string { return Success.Render("✓") }
func MarkFail() string    { return Error.Render("✗") }
func MarkWarn() string    { return Warn.Render("⚠") }
func MarkInfo() string    { return Info.Render("ℹ") }
func MarkDot() string     { return Muted.Render("·") }

var plainMode bool

// SetPlain disables ANSI styling. Useful for CI snapshots and tests.
func SetPlain(plain bool) {
	plainMode = plain
	if !plain {
		return
	}
	noop := lipgloss.NewStyle()
	Heading = noop
	Success = noop
	Warn = noop
	Error = noop
	Muted = noop
	Info = noop
	Accent = noop
}

// IsPlain reports whether ANSI styling is currently disabled.
func IsPlain() bool { return plainMode }
