// SPDX-License-Identifier: MIT

// Package watchcmd implements `harness logs --follow`: a poll-based JSONL
// tail that renders as the spec §6 TUI panel. Pure Bubble Tea — no external
// process management; safe to Ctrl-C.
package watchcmd

import (
	"bufio"
	stdctx "context"
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Path     string
	Tail     int           // initial lines to show
	Interval time.Duration // poll interval
}

func Run(_ stdctx.Context, opts Options, out io.Writer) error {
	if opts.Path == "" {
		return fmt.Errorf("watch: empty path")
	}
	if opts.Interval <= 0 {
		opts.Interval = 750 * time.Millisecond
	}
	if opts.Tail <= 0 {
		opts.Tail = 20
	}
	m := model{path: opts.Path, interval: opts.Interval, tail: opts.Tail}
	prog := tea.NewProgram(m, tea.WithOutput(out))
	_, err := prog.Run()
	return err
}

type model struct {
	path     string
	interval time.Duration
	tail     int
	lines    []string
	offset   int64
	err      error
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Batch(loadInitial(m.path, m.tail), tick(m.interval))
}

type initLoadedMsg struct {
	lines  []string
	offset int64
	err    error
}

type pollResultMsg struct {
	newLines []string
	offset   int64
	err      error
}

func loadInitial(path string, tail int) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return initLoadedMsg{err: err}
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
		var all []string
		for scanner.Scan() {
			all = append(all, scanner.Text())
		}
		end, _ := f.Seek(0, io.SeekCurrent)
		if tail > 0 && len(all) > tail {
			all = all[len(all)-tail:]
		}
		return initLoadedMsg{lines: all, offset: end}
	}
}

func poll(path string, from int64) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return pollResultMsg{err: err, offset: from}
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return pollResultMsg{err: err, offset: from}
		}
		if info.Size() < from {
			// file rotated; reset
			from = 0
		}
		if _, err := f.Seek(from, io.SeekStart); err != nil {
			return pollResultMsg{err: err, offset: from}
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
		var add []string
		for scanner.Scan() {
			add = append(add, scanner.Text())
		}
		end, _ := f.Seek(0, io.SeekCurrent)
		return pollResultMsg{newLines: add, offset: end}
	}
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case initLoadedMsg:
		m.err = v.err
		m.lines = v.lines
		m.offset = v.offset
	case pollResultMsg:
		m.err = v.err
		if len(v.newLines) > 0 {
			m.lines = append(m.lines, v.newLines...)
			if len(m.lines) > 200 {
				m.lines = m.lines[len(m.lines)-200:]
			}
		}
		m.offset = v.offset
	case tickMsg:
		return m, tea.Batch(poll(m.path, m.offset), tick(m.interval))
	case tea.KeyMsg:
		switch v.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

var (
	header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4338CA"))
	muted  = lipgloss.NewStyle().Faint(true)
	errSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("#b91c1c"))
)

func (m model) View() string {
	s := header.Render("HarnessX — logs --follow") + "  " + muted.Render("(q to quit)") + "\n\n"
	if m.err != nil {
		return s + errSt.Render("error: "+m.err.Error()) + "\n"
	}
	if len(m.lines) == 0 {
		return s + muted.Render("no log entries yet…") + "\n"
	}
	for _, line := range m.lines {
		s += line + "\n"
	}
	return s
}
