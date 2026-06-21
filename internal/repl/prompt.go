// SPDX-License-Identifier: MIT

package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

// promptReader abstracts over the two ways the chat REPL collects a
// line: chzyer/readline when stdin is a TTY (gives up/down history +
// TAB completion), and a plain bufio.Reader otherwise (smoke scripts,
// piped input). The interface keeps repl.Run free of the readline
// dependency check at every callsite.
type promptReader interface {
	ReadInput(prompt, continuation string) (string, error)
	Close() error
}

// newPromptReader builds the right reader for the given in/out. The
// readline path is taken only when both stdin and stdout look like a
// TTY (file descriptors check). historyPath persists between sessions
// when non-empty; a missing directory is created on demand. completer
// supplies the static completion candidates (slash commands, adapter
// ids, session labels) when TAB is pressed.
func newPromptReader(in io.Reader, out io.Writer, historyPath string, completer readline.AutoCompleter) promptReader {
	if !isTerminal(in) || !isTerminal(out) {
		return &bufioPromptReader{r: bufio.NewReader(in), w: out}
	}
	if historyPath != "" {
		_ = os.MkdirAll(filepath.Dir(historyPath), 0o755)
	}
	cfg := &readline.Config{
		HistoryFile:            historyPath,
		HistorySearchFold:      true,
		AutoComplete:           completer,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		DisableAutoSaveHistory: false,
		Stdin:                  asReadCloser(in),
		Stdout:                 out,
		Stderr:                 out,
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		// Fall back to bufio when readline init fails (e.g. exotic
		// terminal). Always recoverable.
		return &bufioPromptReader{r: bufio.NewReader(in), w: out}
	}
	return &readlinePromptReader{rl: rl}
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
	var parts []string
	for {
		line, err := b.r.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		trimmed := strings.TrimRight(line, "\n")
		if strings.HasSuffix(trimmed, "\\") {
			parts = append(parts, strings.TrimSuffix(trimmed, "\\"))
			_, _ = fmt.Fprint(b.w, continuation)
			continue
		}
		parts = append(parts, trimmed)
		break
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func (b *bufioPromptReader) Close() error { return nil }

type readlinePromptReader struct {
	rl *readline.Instance
}

func (r *readlinePromptReader) ReadInput(prompt, continuation string) (string, error) {
	r.rl.SetPrompt(prompt)
	var parts []string
	for {
		line, err := r.rl.Readline()
		if err != nil {
			// Ctrl-C at empty prompt → treat as a soft cancel, return
			// empty so the caller re-prompts. Ctrl-D → io.EOF, lets
			// the caller exit cleanly.
			if err == readline.ErrInterrupt {
				if len(parts) > 0 {
					parts = parts[:0]
					r.rl.SetPrompt(prompt)
					continue
				}
				return "", nil
			}
			return "", err
		}
		trimmed := strings.TrimRight(line, "\n")
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

func (r *readlinePromptReader) Close() error { return r.rl.Close() }

// chatCompleter builds a TAB-completion tree over the static slash
// command list plus the dynamic adapter ids and saved session labels
// so users get hints without having to remember the cheatsheet.
func chatCompleter(adapterIDs, labels []string) readline.AutoCompleter {
	staticSlashes := []string{
		"/help", "/exit", "/quit", "/history", "/last",
		"/agents", "/cost", "/diff", "/clear", "/timeline",
		"/auto-gate on", "/auto-gate off",
		"/save", "/branch", "/recap",
		"/save-prompt", "/prompt", "/prompts",
		"/exec", "/do", "/ship", "/ci", "/test", "/lint",
		"/budget", "/goal", "/plan",
	}
	items := make([]readline.PrefixCompleterInterface, 0, len(staticSlashes)+2)
	for _, s := range staticSlashes {
		items = append(items, readline.PcItem(s))
	}
	if len(adapterIDs) > 0 {
		dyn := make([]readline.PrefixCompleterInterface, 0, len(adapterIDs))
		for _, a := range adapterIDs {
			dyn = append(dyn, readline.PcItem(a))
		}
		items = append(items, readline.PcItem("/use", dyn...))
	}
	if len(labels) > 0 {
		dyn := make([]readline.PrefixCompleterInterface, 0, len(labels))
		for _, l := range labels {
			dyn = append(dyn, readline.PcItem(l))
		}
		items = append(items, readline.PcItem("/resume", dyn...))
	}
	return readline.NewPrefixCompleter(items...)
}
