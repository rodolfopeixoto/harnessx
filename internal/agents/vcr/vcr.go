// SPDX-License-Identifier: MIT

// Package vcr wraps an AgentAdapter to record real CLI output on the
// first call for a given prompt fingerprint and replay it on every
// subsequent call. Lets E2E tests exercise the real wire format
// (Claude JSON-Lines, codex exec stream) without burning tokens on
// every CI run. Fingerprint = SHA-1(prompt + model + workingDir
// basename), so a tiny prompt change forces a fresh recording.
package vcr

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type Mode int

const (
	ModeAuto   Mode = iota // replay if exists, record otherwise
	ModeReplay             // require recording; error if missing
	ModeRecord             // always re-record
)

type Adapter struct {
	inner agents.AgentAdapter
	dir   string
	mode  Mode
	now   func() time.Time
}

type Options struct {
	Inner agents.AgentAdapter
	Dir   string
	Mode  Mode
}

func New(opts Options) *Adapter {
	return &Adapter{
		inner: opts.Inner,
		dir:   opts.Dir,
		mode:  opts.Mode,
		now:   time.Now,
	}
}

func (a *Adapter) ID() string                        { return a.inner.ID() }
func (a *Adapter) Name() string                      { return "vcr(" + a.inner.Name() + ")" }
func (a *Adapter) Capabilities() agents.Capabilities { return a.inner.Capabilities() }

func (a *Adapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	return a.inner.Healthcheck(ctx)
}

func (a *Adapter) ParseUsage(o agents.AgentOutput) agents.Usage {
	return a.inner.ParseUsage(o)
}

func (a *Adapter) ClassifyFailure(o agents.AgentOutput, code int, err error) agents.FailureType {
	return a.inner.ClassifyFailure(o, code, err)
}

func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	fp := fingerprint(req, a.inner.ID())
	path := filepath.Join(a.dir, fp+".json")

	if a.mode != ModeRecord {
		if cached, err := load(path); err == nil {
			cached.replayedAt = a.now()
			return cached.toResult()
		} else if a.mode == ModeReplay {
			return agents.AgentResult{
				Err: fmt.Errorf("vcr: replay missing for fingerprint %s (%s)", fp, path),
			}
		}
	}

	res := a.inner.Run(ctx, req)
	cassette := fromResult(res, fp, a.inner.ID())
	if err := save(path, cassette); err != nil {
		res.Err = errors.Join(res.Err, fmt.Errorf("vcr: save cassette: %w", err))
	}
	return res
}

type cassette struct {
	Fingerprint  string       `json:"fingerprint"`
	AdapterID    string       `json:"adapter_id"`
	RecordedAt   time.Time    `json:"recorded_at"`
	ExitCode     int          `json:"exit_code"`
	DurationMS   int64        `json:"duration_ms"`
	Stdout       []byte       `json:"stdout,omitempty"`
	Stderr       []byte       `json:"stderr,omitempty"`
	FinalMessage string       `json:"final_message,omitempty"`
	Usage        agents.Usage `json:"usage"`
	Failure      string       `json:"failure"`
	ErrMessage   string       `json:"err,omitempty"`

	replayedAt time.Time
}

func (c cassette) toResult() agents.AgentResult {
	res := agents.AgentResult{
		Output: agents.AgentOutput{
			Stdout: c.Stdout, Stderr: c.Stderr, FinalMessage: c.FinalMessage,
		},
		ExitCode: c.ExitCode,
		Duration: time.Duration(c.DurationMS) * time.Millisecond,
		Usage:    c.Usage,
		Failure:  agents.FailureType(c.Failure),
	}
	if c.ErrMessage != "" {
		res.Err = errors.New(c.ErrMessage)
	}
	return res
}

func fromResult(res agents.AgentResult, fp, id string) cassette {
	c := cassette{
		Fingerprint: fp, AdapterID: id,
		RecordedAt: time.Now().UTC(),
		ExitCode:   res.ExitCode,
		DurationMS: res.Duration.Milliseconds(),
		Stdout:     res.Output.Stdout, Stderr: res.Output.Stderr,
		FinalMessage: res.Output.FinalMessage,
		Usage:        res.Usage,
		Failure:      string(res.Failure),
	}
	if res.Err != nil {
		c.ErrMessage = res.Err.Error()
	}
	return c
}

func fingerprint(req agents.AgentRequest, adapterID string) string {
	h := sha1.New()
	h.Write([]byte(adapterID))
	h.Write([]byte{0})
	h.Write([]byte(req.Model))
	h.Write([]byte{0})
	h.Write([]byte(filepath.Base(req.WorkingDir)))
	h.Write([]byte{0})
	h.Write([]byte(req.Prompt))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func load(path string) (cassette, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return cassette{}, err
	}
	var c cassette
	if err := json.Unmarshal(body, &c); err != nil {
		return cassette{}, err
	}
	return c, nil
}

func save(path string, c cassette) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
