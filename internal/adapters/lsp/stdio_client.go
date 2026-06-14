// SPDX-License-Identifier: MIT

package lsp

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
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

func parseDocumentSymbols(raw json.RawMessage, path string) []Symbol {
	var probe []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil || len(probe) == 0 {
		return nil
	}
	_, isFlat := probe[0]["location"]
	if isFlat {
		var flat []struct {
			Name     string `json:"name"`
			Location struct {
				Range struct {
					Start struct{ Line int } `json:"start"`
				} `json:"range"`
			} `json:"location"`
		}
		if err := json.Unmarshal(raw, &flat); err != nil {
			return nil
		}
		out := make([]Symbol, 0, len(flat))
		for _, s := range flat {
			out = append(out, Symbol{Name: s.Name, Path: path, Line: s.Location.Range.Start.Line + 1})
		}
		return out
	}
	var hier []struct {
		Name  string `json:"name"`
		Range struct {
			Start struct{ Line int } `json:"start"`
		} `json:"range"`
		Children []struct {
			Name  string `json:"name"`
			Range struct {
				Start struct{ Line int } `json:"start"`
			} `json:"range"`
		} `json:"children"`
	}
	if err := json.Unmarshal(raw, &hier); err != nil {
		return nil
	}
	out := make([]Symbol, 0, len(hier))
	for _, s := range hier {
		out = append(out, Symbol{Name: s.Name, Path: path, Line: s.Range.Start.Line + 1})
		for _, c := range s.Children {
			out = append(out, Symbol{Name: s.Name + "." + c.Name, Path: path, Line: c.Range.Start.Line + 1})
		}
	}
	return out
}

func severityName(s int) string {
	switch s {
	case 1:
		return "error"
	case 2:
		return "warning"
	case 3:
		return "information"
	case 4:
		return "hint"
	}
	return "info"
}

func pathToURI(p string) string {
	abs, _ := filepath.Abs(p)
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}

func uriToPath(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	return parsed.Path
}

func repoHash(root string) string {
	sum := sha256.Sum256([]byte(root))
	return hex.EncodeToString(sum[:8])
}

func queryHash(method, path string) string {
	sum := sha256.Sum256([]byte(method + "|" + path))
	return hex.EncodeToString(sum[:16])
}

type cachedResult struct {
	Symbols     []Symbol     `json:"symbols,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

func readCache(path string) (cachedResult, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return cachedResult{}, false
	}
	var c cachedResult
	if err := json.Unmarshal(b, &c); err != nil {
		return cachedResult{}, false
	}
	return c, true
}

func writeCache(path string, c cachedResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

type rawResponse struct {
	Result json.RawMessage
	Err    *rpcError
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
