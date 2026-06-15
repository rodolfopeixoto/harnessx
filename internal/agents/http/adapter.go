// SPDX-License-Identifier: MIT

// Package http wraps an HTTP/JSON API endpoint as an AgentAdapter.
// Used for direct provider calls (Anthropic, OpenAI, Gemini, Moonshot,
// minimax) without spawning a CLI binary. Secrets resolved via
// internal/secrets.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/yaml"
	"github.com/ropeixoto/harnessx/internal/secrets"
)

type Adapter struct {
	Spec    yaml.Spec
	Secrets *secrets.Store
	Client  *http.Client
}

func New(spec yaml.Spec) *Adapter {
	return &Adapter{
		Spec:    spec,
		Secrets: secrets.New(),
		Client:  &http.Client{},
	}
}

func (a *Adapter) ID() string                        { return a.Spec.ID }
func (a *Adapter) Name() string                      { return a.Spec.Name }
func (a *Adapter) Capabilities() agents.Capabilities { return a.Spec.Capabilities }

func (a *Adapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	if _, err := a.resolveSecret(); err != nil {
		return agents.HealthcheckResult{OK: false, Err: "missing credential: " + err.Error()}
	}
	return agents.HealthcheckResult{OK: true, Detail: "secret resolved; live ping skipped (no /health endpoint)"}
}

func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	start := time.Now()
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = time.Duration(maxInt(a.Spec.API.TimeoutSeconds, 60)) * time.Second
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, err := a.buildBody(req)
	if err != nil {
		return agents.AgentResult{Err: err, Failure: agents.FailurePermanent, Duration: time.Since(start)}
	}
	httpReq, err := http.NewRequestWithContext(rctx, methodOrPost(a.Spec.API.Method), a.Spec.API.Endpoint, bytes.NewReader(body))
	if err != nil {
		return agents.AgentResult{Err: err, Failure: agents.FailurePermanent, Duration: time.Since(start)}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range a.Spec.API.Headers {
		httpReq.Header.Set(k, v)
	}
	if err := a.applyAuth(httpReq); err != nil {
		return agents.AgentResult{Err: err, Failure: agents.FailureAuth, Duration: time.Since(start)}
	}

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return agents.AgentResult{Err: err, Failure: agents.FailureTransient, Duration: time.Since(start)}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return agents.AgentResult{Err: err, Failure: agents.FailureTransient, Duration: time.Since(start)}
	}

	out := agents.AgentOutput{Stdout: respBody}
	out.FinalMessage = extractJSONPath(respBody, a.Spec.API.Response.FinalMessage)

	res := agents.AgentResult{
		Output:   out,
		ExitCode: resp.StatusCode,
		Duration: time.Since(start),
	}
	if resp.StatusCode >= 400 {
		res.Err = fmt.Errorf("http %d: %s", resp.StatusCode, truncateForErr(string(respBody), 200))
		res.Failure = classifyHTTP(resp.StatusCode)
	}
	res.Usage = a.ParseUsage(out)
	return res
}

func (a *Adapter) ParseUsage(o agents.AgentOutput) agents.Usage {
	usage := agents.Usage{}
	if a.Spec.API.Response.Usage == "" {
		return usage
	}
	raw := extractJSONPath(o.Stdout, a.Spec.API.Response.Usage)
	if raw == "" {
		return usage
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return usage
	}
	usage.InputTokens = intFromAny(parsed, "input_tokens", "prompt_tokens")
	usage.OutputTokens = intFromAny(parsed, "output_tokens", "completion_tokens")
	usage.CachedInputTokens = intFromAny(parsed, "cached_input_tokens", "cache_read_input_tokens")
	usage.Mode = "reported"
	return usage
}

func (a *Adapter) ClassifyFailure(_ agents.AgentOutput, exitCode int, runErr error) agents.FailureType {
	if runErr == nil && exitCode < 400 {
		return agents.FailureNone
	}
	return classifyHTTP(exitCode)
}

func (a *Adapter) buildBody(req agents.AgentRequest) ([]byte, error) {
	tmpl := a.Spec.API.RequestTemplate
	if tmpl == "" {
		return nil, errors.New("http: empty request_template")
	}
	model := req.Model
	if model == "" {
		if a.Spec.Capabilities.Models != nil {
			model = a.Spec.Capabilities.Models["default"]
		}
	}
	out := tmpl
	out = strings.ReplaceAll(out, "{{prompt}}", jsonString(req.Prompt))
	out = strings.ReplaceAll(out, "{{model}}", model)
	for k, v := range req.Extra {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return []byte(out), nil
}

func (a *Adapter) applyAuth(req *http.Request) error {
	if a.Spec.API.Auth.SecretRef == "" {
		return nil
	}
	value, err := a.resolveSecret()
	if err != nil {
		return err
	}
	header := a.Spec.API.Auth.Header
	if header == "" {
		header = "Authorization"
	}
	scheme := a.Spec.API.Auth.Scheme
	if scheme != "" {
		req.Header.Set(header, scheme+" "+value)
	} else {
		req.Header.Set(header, value)
	}
	return nil
}

func (a *Adapter) resolveSecret() (string, error) {
	if a.Spec.API.Auth.SecretRef == "" {
		return "", nil
	}
	return a.Secrets.Resolve(a.Spec.API.Auth.SecretRef)
}

func methodOrPost(m string) string {
	if m == "" {
		return http.MethodPost
	}
	return strings.ToUpper(m)
}

func classifyHTTP(code int) agents.FailureType {
	switch {
	case code == 401 || code == 403:
		return agents.FailureAuth
	case code == 429:
		return agents.FailureRateLimit
	case code == 408 || code == 504:
		return agents.FailureTimeout
	case code >= 500:
		return agents.FailureTransient
	case code >= 400:
		return agents.FailurePermanent
	default:
		return agents.FailureNone
	}
}

func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return strings.Trim(string(b), `"`)
}

func extractJSONPath(data []byte, path string) string {
	if len(data) == 0 || path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return ""
	}
	cur := root
	for _, p := range parts {
		switch v := cur.(type) {
		case map[string]any:
			cur = v[p]
		case []any:
			if idx, ok := arrayIndex(p); ok && idx >= 0 && idx < len(v) {
				cur = v[idx]
			} else {
				return ""
			}
		default:
			return ""
		}
	}
	if cur == nil {
		return ""
	}
	if s, ok := cur.(string); ok {
		return s
	}
	out, _ := json.Marshal(cur)
	return string(out)
}

func arrayIndex(s string) (int, bool) {
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = strings.TrimSuffix(strings.TrimPrefix(s, "["), "]")
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, false
	}
	return n, true
}

func intFromAny(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if f, ok := v.(float64); ok {
				return int(f)
			}
		}
	}
	return 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateForErr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
