// SPDX-License-Identifier: MIT

package interactive

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/creack/pty"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type PTYStrategy struct{}

func (PTYStrategy) ID() string { return "pty" }

func (PTYStrategy) Run(ctx context.Context, cfg Config, req agents.AgentRequest) (agents.AgentResult, error) {
	if cfg.Binary == "" {
		return agents.AgentResult{}, errors.New("interactive: pty: binary required")
	}
	timeout := durationOrDefault(cfg.HardTimeoutSeconds, defaultHardTimeout)
	idle := idleThreshold(cfg.IdleMs)
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(rctx, cfg.Binary, cfg.Args...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return agents.AgentResult{}, err
	}
	defer func() { _ = ptmx.Close() }()
	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: 40, Cols: 120})

	collector := newIdleCollector(idle, cfg.BannerPattern)
	start := time.Now()

	if err := waitForReady(rctx, ptmx, collector); err != nil {
		_ = cmd.Process.Kill()
		return failedResult(err, start), err
	}

	prompt := promptFromRequest(req)
	if _, err := io.WriteString(ptmx, prompt+"\n"); err != nil {
		_ = cmd.Process.Kill()
		return failedResult(err, start), err
	}

	collector.reset()
	if err := collector.collectUntilIdle(rctx, ptmx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		_ = cmd.Process.Kill()
		return failedResult(err, start), err
	}

	_, _ = io.WriteString(ptmx, "/exit\n")
	_ = cmd.Wait()

	body := collector.bytesAfterFirstPrompt()
	return agents.AgentResult{
		Output:   agents.AgentOutput{Stdout: body, FinalMessage: string(body)},
		Duration: time.Since(start),
	}, nil
}

type idleCollector struct {
	mu            sync.Mutex
	buf           bytes.Buffer
	bannerSeen    bool
	bannerOffset  int
	idle          time.Duration
	bannerRegex   *regexp.Regexp
	lastWriteAtNs int64
}

func newIdleCollector(idle time.Duration, banner string) *idleCollector {
	c := &idleCollector{idle: idle}
	if banner != "" {
		if re, err := regexp.Compile(banner); err == nil {
			c.bannerRegex = re
		}
	}
	return c
}

func (c *idleCollector) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bannerOffset = c.buf.Len()
	c.lastWriteAtNs = time.Now().UnixNano()
}

func (c *idleCollector) write(p []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf.Write(p)
	c.lastWriteAtNs = time.Now().UnixNano()
}

func (c *idleCollector) hasBanner() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bannerRegex == nil {
		return c.buf.Len() > 0
	}
	if c.bannerRegex.Match(c.buf.Bytes()) {
		c.bannerSeen = true
	}
	return c.bannerSeen
}

func (c *idleCollector) idleFor() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastWriteAtNs == 0 {
		return 0
	}
	return time.Duration(time.Now().UnixNano() - c.lastWriteAtNs)
}

func (c *idleCollector) bytesAfterFirstPrompt() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.bannerOffset > c.buf.Len() {
		return nil
	}
	return append([]byte(nil), c.buf.Bytes()[c.bannerOffset:]...)
}

func (c *idleCollector) collectUntilIdle(ctx context.Context, r io.Reader) error {
	buf := make([]byte, 4096)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_ = setReadDeadline(r, time.Now().Add(c.idle))
		n, err := r.Read(buf)
		if n > 0 {
			c.write(buf[:n])
			continue
		}
		if err == nil {
			continue
		}
		if isTimeoutErr(err) {
			if c.idleFor() >= c.idle {
				return nil
			}
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
}

func waitForReady(ctx context.Context, r io.Reader, c *idleCollector) error {
	buf := make([]byte, 4096)
	deadline := time.Now().Add(15 * time.Second)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return errors.New("interactive: pty: REPL ready prompt not seen in 15s")
		}
		_ = setReadDeadline(r, time.Now().Add(500*time.Millisecond))
		n, err := r.Read(buf)
		if n > 0 {
			c.write(buf[:n])
		}
		if c.hasBanner() {
			return nil
		}
		if err == nil {
			continue
		}
		if isTimeoutErr(err) {
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
}

func promptFromRequest(req agents.AgentRequest) string {
	if req.Stdin != "" {
		return req.Stdin
	}
	return req.Prompt
}

func failedResult(err error, start time.Time) agents.AgentResult {
	return agents.AgentResult{
		Err:      err,
		Failure:  agents.FailureTransient,
		Duration: time.Since(start),
	}
}

func durationOrDefault(seconds, fallback int) time.Duration {
	if seconds <= 0 {
		return time.Duration(fallback) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func idleThreshold(ms int) time.Duration {
	if ms <= 0 {
		return defaultIdle
	}
	return time.Duration(ms) * time.Millisecond
}

const (
	defaultIdle        = 1500 * time.Millisecond
	defaultHardTimeout = 180
)
