// SPDX-License-Identifier: MIT

package containers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Compose struct {
	Binary     string
	File       string
	WorkingDir string
}

func NewCompose(file string) Compose {
	return Compose{Binary: constants.DefaultDockerBinary, File: file}
}

func (c Compose) binary() string {
	if c.Binary != "" {
		return c.Binary
	}
	return constants.DefaultDockerBinary
}

func (c Compose) Up(ctx context.Context) error {
	return c.run(ctx, "up", "-d", "--remove-orphans")
}

func (c Compose) Down(ctx context.Context) error {
	return c.run(ctx, "down", "--remove-orphans", "--volumes")
}

func (c Compose) run(ctx context.Context, args ...string) error {
	full := append([]string{"compose", "-f", c.File}, args...)
	ctx, cancel := context.WithTimeout(ctx, constants.DefaultContainerUpTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.binary(), full...)
	cmd.Dir = c.WorkingDir
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("containers compose %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(errBuf.String()))
	}
	return nil
}

type HealthProbe struct {
	URL     string
	Client  *http.Client
	Timeout time.Duration
	Backoff time.Duration
}

func NewHealthProbe(url string) HealthProbe {
	return HealthProbe{
		URL:     url,
		Client:  &http.Client{Timeout: 2 * time.Second},
		Timeout: 30 * time.Second,
		Backoff: 250 * time.Millisecond,
	}
}

func (p HealthProbe) Wait(ctx context.Context) error {
	deadline := time.Now().Add(p.Timeout)
	for {
		if err := p.probeOnce(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("containers: health probe %s timeout", p.URL)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.Backoff):
		}
	}
}

func (p HealthProbe) probeOnce(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return err
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("containers: probe status %d", resp.StatusCode)
}

func VerifyClean(ctx context.Context, lister Lister) error {
	items, err := lister.List(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	names := make([]string, 0, len(items))
	for _, i := range items {
		names = append(names, i.Name)
	}
	return fmt.Errorf("containers: %d container(s) still present: %s", len(items), strings.Join(names, ","))
}

var ErrDockerMissing = errors.New("containers: docker binary not on PATH")
