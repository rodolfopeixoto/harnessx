// SPDX-License-Identifier: MIT

package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Stdio is a minimal LSP client that talks to a language server over its
// stdin/stdout. Most servers (gopls, ruby-lsp, pyright, rust-analyzer,
// typescript-language-server) plug in by setting the binary + args; quirks
// land via per-server wrappers.
type Stdio struct {
	binary     string
	args       []string
	language   string
	languageID string

	root   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	cancel context.CancelFunc

	mu      sync.Mutex
	pending map[int]chan rawResponse
	diags   map[string][]Diagnostic
	nextID  int64
	started bool
	closed  bool
}

// NewStdio constructs a Stdio client. Call Start before use.
func NewStdio(binary string, args []string, language, languageID, root string) *Stdio {
	return &Stdio{
		binary: binary, args: args,
		language: language, languageID: languageID, root: root,
		pending: map[int]chan rawResponse{},
		diags:   map[string][]Diagnostic{},
	}
}

func (s *Stdio) Language() string { return s.language }

func (s *Stdio) Start(ctx context.Context) error {
	if _, err := exec.LookPath(s.binary); err != nil {
		return fmt.Errorf("lsp/%s: binary %q not on PATH", s.language, s.binary)
	}
	rctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.cmd = exec.CommandContext(rctx, s.binary, s.args...)
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		cancel()
		return err
	}
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	s.cmd.Stderr = io.Discard
	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("lsp/%s: start: %w", s.language, err)
	}
	s.stdin = stdin
	s.stdout = bufio.NewReader(stdout)
	go s.readLoop()

	initParams := map[string]any{
		"processId":             os.Getpid(),
		"rootUri":               pathToURI(s.root),
		"capabilities":          map[string]any{},
		"initializationOptions": map[string]any{},
	}
	hctx, hcancel := context.WithTimeout(ctx, 15*time.Second)
	defer hcancel()
	if _, err := s.call(hctx, "initialize", initParams); err != nil {
		_ = s.Close()
		return fmt.Errorf("lsp/%s: initialize: %w", s.language, err)
	}
	if err := s.notify("initialized", map[string]any{}); err != nil {
		_ = s.Close()
		return err
	}
	s.started = true
	return nil
}

func (s *Stdio) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	if s.started {
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, _ = s.call(sctx, "shutdown", nil)
		_ = s.notify("exit", nil)
		cancel()
	}
	if s.cancel != nil {
		s.cancel()
	}
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	return nil
}

func (s *Stdio) DocumentSymbols(ctx context.Context, root, path string) ([]Symbol, bool, error) {
	key := queryHash("documentSymbol", path)
	cachePath := CacheKey(root, repoHash(root), s.language, key)
	if cached, ok := readCache(cachePath); ok {
		return cached.Symbols, true, nil
	}
	abs := filepath.Join(root, path)
	body, err := os.ReadFile(abs)
	if err != nil {
		return nil, false, err
	}
	if err := s.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri": pathToURI(abs), "languageId": s.languageID, "version": 1,
			"text": string(body),
		},
	}); err != nil {
		return nil, false, err
	}
	raw, err := s.call(ctx, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": pathToURI(abs)},
	})
	if err != nil {
		return nil, false, err
	}
	syms := parseDocumentSymbols(raw, path)
	_ = writeCache(cachePath, cachedResult{Symbols: syms})
	return syms, false, nil
}

func (s *Stdio) Diagnostics(ctx context.Context, root, path string) ([]Diagnostic, bool, error) {
	key := queryHash("diagnostics", path)
	cachePath := CacheKey(root, repoHash(root), s.language, key)
	if cached, ok := readCache(cachePath); ok {
		return cached.Diagnostics, true, nil
	}
	abs := filepath.Join(root, path)
	body, err := os.ReadFile(abs)
	if err != nil {
		return nil, false, err
	}
	if err := s.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri": pathToURI(abs), "languageId": s.languageID, "version": 1,
			"text": string(body),
		},
	}); err != nil {
		return nil, false, err
	}
	deadline := time.Now().Add(2 * time.Second)
	uri := pathToURI(abs)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		out, ok := s.diags[uri]
		s.mu.Unlock()
		if ok {
			_ = writeCache(cachePath, cachedResult{Diagnostics: out})
			return out, false, nil
		}
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}
	_ = writeCache(cachePath, cachedResult{})
	return nil, false, nil
}

func (s *Stdio) Definitions(ctx context.Context, root, path string, line, col int) ([]Symbol, bool, error) {
	return nil, false, nil
}
func (s *Stdio) References(ctx context.Context, root, path string, line, col int) ([]Symbol, bool, error) {
	return nil, false, nil
}

// --- LSP protocol --------------------------------------------------------

func (s *Stdio) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	id := int(atomic.AddInt64(&s.nextID, 1))
	ch := make(chan rawResponse, 1)
	s.mu.Lock()
	s.pending[id] = ch
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()
	if err := s.send(map[string]any{
		"jsonrpc": "2.0", "id": id, "method": method, "params": params,
	}); err != nil {
		return nil, err
	}
	select {
	case r := <-ch:
		if r.Err != nil {
			return nil, fmt.Errorf("lsp/%s: %s: %s", s.language, method, r.Err.Message)
		}
		return r.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Stdio) notify(method string, params any) error {
	return s.send(map[string]any{
		"jsonrpc": "2.0", "method": method, "params": params,
	})
}

func (s *Stdio) send(msg any) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(b))
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("lsp: closed")
	}
	if _, err := s.stdin.Write([]byte(header)); err != nil {
		return err
	}
	_, err = s.stdin.Write(b)
	return err
}

func (s *Stdio) readLoop() {
	for {
		msg, err := readMessage(s.stdout)
		if err != nil {
			return
		}
		var probe struct {
			ID     *int            `json:"id"`
			Method string          `json:"method"`
			Result json.RawMessage `json:"result"`
			Error  *rpcError       `json:"error"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(msg, &probe); err != nil {
			continue
		}
		if probe.ID != nil && probe.Method == "" {
			s.mu.Lock()
			ch, ok := s.pending[*probe.ID]
			s.mu.Unlock()
			if ok {
				ch <- rawResponse{Result: probe.Result, Err: probe.Error}
			}
			continue
		}
		if probe.Method == "textDocument/publishDiagnostics" {
			var pub struct {
				URI         string `json:"uri"`
				Diagnostics []struct {
					Range struct {
						Start struct{ Line, Character int } `json:"start"`
					} `json:"range"`
					Severity int    `json:"severity"`
					Message  string `json:"message"`
				} `json:"diagnostics"`
			}
			if err := json.Unmarshal(probe.Params, &pub); err == nil {
				out := make([]Diagnostic, 0, len(pub.Diagnostics))
				for _, d := range pub.Diagnostics {
					out = append(out, Diagnostic{
						Path:     uriToPath(pub.URI),
						Line:     d.Range.Start.Line + 1,
						Severity: severityName(d.Severity),
						Message:  d.Message,
					})
				}
				s.mu.Lock()
				s.diags[pub.URI] = out
				s.mu.Unlock()
			}
		}
	}
}

// --- shared helpers (used by every wrapper) -----------------------------

func readMessage(r *bufio.Reader) ([]byte, error) {
	var length int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			n, err := strconv.Atoi(strings.TrimSpace(line[len("Content-Length:"):]))
			if err != nil {
				return nil, err
			}
			length = n
		}
	}
	if length == 0 {
		return nil, errors.New("lsp: missing Content-Length")
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

type rawResponse struct {
	Result json.RawMessage
	Err    *rpcError
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
