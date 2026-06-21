// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

type promptReader interface {
	ReadInput(prompt, continuation string) (string, error)
	Close() error
}

var staticSlashCommands = []string{
	"/help", "/exit", "/quit", "/history", "/last",
	"/agents", "/cost", "/diff", "/clear", "/timeline",
	"/auto-gate on", "/auto-gate off",
	"/save", "/branch", "/recap",
	"/save-prompt", "/prompt", "/prompts",
	"/exec", "/do", "/ship", "/drive", "/ci", "/test", "/lint",
	"/budget", "/goal", "/plan",
}

func newPromptReader(in io.Reader, out io.Writer, historyPath string, completer readline.AutoCompleter) promptReader {
	if !isTerminal(in) || !isTerminal(out) {
		return &bufioPromptReader{r: bufio.NewReader(in), w: out}
	}
	if historyPath != "" {
		_ = os.MkdirAll(filepath.Dir(historyPath), 0o755)
	}
	paste := newPasteCoalescingReader(asReadCloser(in))
	cfg := &readline.Config{
		HistoryFile:            historyPath,
		HistorySearchFold:      true,
		AutoComplete:           completer,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		DisableAutoSaveHistory: false,
		Stdin:                  paste,
		Stdout:                 out,
		Stderr:                 out,
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		return &bufioPromptReader{r: bufio.NewReader(in), w: out}
	}
	_, _ = io.WriteString(out, bracketedPasteEnableSeq)
	r := &readlinePromptReader{rl: rl, paste: paste, out: out}
	r.startWinch()
	return r
}

func isTerminal(x interface{}) bool {
	type fileLike interface {
		Fd() uintptr
	}
	f, ok := x.(fileLike)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func asReadCloser(r io.Reader) io.ReadCloser {
	if rc, ok := r.(io.ReadCloser); ok {
		return rc
	}
	return io.NopCloser(r)
}

type bufioPromptReader struct {
	r *bufio.Reader
	w io.Writer
}

func (b *bufioPromptReader) ReadInput(prompt, continuation string) (string, error) {
	if _, err := fmt.Fprint(b.w, prompt); err != nil {
		return "", err
	}
	parts, err := readBackslashContinuation(b.r, b.w, continuation)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func (b *bufioPromptReader) Close() error { return nil }

func readBackslashContinuation(r *bufio.Reader, w io.Writer, continuation string) ([]string, error) {
	var parts []string
	heredoc := false
	for {
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			return nil, err
		}
		trimmed := strings.TrimRight(line, "\n")
		if heredoc {
			if strings.TrimSpace(trimmed) == hereDocMarker {
				return parts, nil
			}
			parts = append(parts, trimmed)
			_, _ = fmt.Fprint(w, continuation)
			continue
		}
		if strings.TrimSpace(trimmed) == hereDocMarker {
			heredoc = true
			_, _ = fmt.Fprint(w, continuation)
			continue
		}
		if strings.HasSuffix(trimmed, "\\") {
			parts = append(parts, strings.TrimSuffix(trimmed, "\\"))
			_, _ = fmt.Fprint(w, continuation)
			continue
		}
		parts = append(parts, trimmed)
		return parts, nil
	}
}

const hereDocMarker = `"""`

type readlinePromptReader struct {
	rl     *readline.Instance
	paste  *pasteCoalescingReader
	out    io.Writer
	winch  chan os.Signal
	closed chan struct{}
}

func (r *readlinePromptReader) startWinch() {
	r.winch = make(chan os.Signal, 1)
	r.closed = make(chan struct{})
	signal.Notify(r.winch, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-r.closed:
				return
			case <-r.winch:
				if r.rl != nil {
					r.rl.Refresh()
				}
			}
		}
	}()
}

func (r *readlinePromptReader) ReadInput(prompt, continuation string) (string, error) {
	r.rl.SetPrompt(prompt)
	var parts []string
	heredoc := false
	for {
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(parts) > 0 {
					parts = parts[:0]
					heredoc = false
					r.rl.SetPrompt(prompt)
					continue
				}
				return "", nil
			}
			return "", err
		}
		trimmed := strings.TrimRight(line, "\n")
		if heredoc {
			if strings.TrimSpace(trimmed) == hereDocMarker {
				break
			}
			parts = append(parts, trimmed)
			r.rl.SetPrompt(continuation)
			continue
		}
		if strings.TrimSpace(trimmed) == hereDocMarker {
			heredoc = true
			r.rl.SetPrompt(continuation)
			continue
		}
		if strings.HasSuffix(trimmed, "\\") {
			parts = append(parts, strings.TrimSuffix(trimmed, "\\"))
			r.rl.SetPrompt(continuation)
			continue
		}
		parts = append(parts, trimmed)
		break
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func (r *readlinePromptReader) Close() error {
	if r.winch != nil {
		signal.Stop(r.winch)
		close(r.closed)
	}
	if r.out != nil {
		_, _ = io.WriteString(r.out, bracketedPasteDisableSeq)
	}
	return r.rl.Close()
}

func chatCompleter(adapterIDs, labels []string) readline.AutoCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(staticSlashCommands)+2)
	for _, s := range staticSlashCommands {
		items = append(items, readline.PcItem(s))
	}
	if len(adapterIDs) > 0 {
		items = append(items, readline.PcItem("/use", toPcItems(adapterIDs)...))
	}
	if len(labels) > 0 {
		items = append(items, readline.PcItem("/resume", toPcItems(labels)...))
	}
	return readline.NewPrefixCompleter(items...)
}

func toPcItems(values []string) []readline.PrefixCompleterInterface {
	out := make([]readline.PrefixCompleterInterface, 0, len(values))
	for _, v := range values {
		out = append(out, readline.PcItem(v))
	}
	return out
}
