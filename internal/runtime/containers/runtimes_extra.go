// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type OrbStack struct{}

func (OrbStack) ID() string     { return "orbstack" }
func (OrbStack) Binary() string { return "orbctl" }

func (o OrbStack) Available(_ context.Context) bool {
	if _, err := exec.LookPath("orbctl"); err == nil {
		return true
	}
	if _, err := exec.LookPath("orb"); err == nil {
		return true
	}
	return false
}

func (o OrbStack) Version(ctx context.Context) (string, error) {
	bin := orbBinary()
	if bin == "" {
		return "", errors.New("orbstack: binary not on PATH")
	}
	out, err := exec.CommandContext(ctx, bin, "version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (o OrbStack) List(ctx context.Context, opts ListOptions) ([]Container, error) {
	return dockerLikeRuntime{binary: "docker"}.List(ctx, opts)
}

func (o OrbStack) Kill(ctx context.Context, id string) error {
	return dockerLikeRuntime{binary: "docker"}.Kill(ctx, id)
}

func (o OrbStack) Prune(ctx context.Context, opts PruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{id: "orbstack", binary: "docker"}.Prune(ctx, opts)
}

func orbBinary() string {
	if path, err := exec.LookPath("orbctl"); err == nil {
		return path
	}
	if path, err := exec.LookPath("orb"); err == nil {
		return path
	}
	return ""
}

type AppleContainer struct{}

func (AppleContainer) ID() string     { return "apple_container" }
func (AppleContainer) Binary() string { return "container" }

func (a AppleContainer) Available(ctx context.Context) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	bin, err := exec.LookPath("container")
	if err != nil {
		return false
	}
	out, err := exec.CommandContext(ctx, bin, "--version").CombinedOutput()
	if err != nil {
		return false
	}
	if !strings.Contains(strings.ToLower(string(out)), "container") {
		return false
	}
	probe := exec.CommandContext(ctx, bin, "list", "--format", "json")
	if err := probe.Run(); err != nil {
		return false
	}
	return true
}

func (a AppleContainer) Version(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "container", "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (a AppleContainer) List(ctx context.Context, opts ListOptions) ([]Container, error) {
	args := []string{"list", "--format", "json"}
	if opts.All {
		args = append(args, "--all")
	}
	out, err := exec.CommandContext(ctx, "container", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("container list: %w", err)
	}
	return parseDockerJSON(out)
}

func (a AppleContainer) Kill(ctx context.Context, id string) error {
	if out, err := exec.CommandContext(ctx, "container", "delete", "--force", id).CombinedOutput(); err != nil {
		return fmt.Errorf("container delete %s: %w: %s", id, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (a AppleContainer) Prune(ctx context.Context, opts PruneOptions) (PruneResult, error) {
	if !opts.IUnderstand {
		return PruneResult{}, errors.New("containers: prune requires IUnderstand=true")
	}
	listed, err := a.List(ctx, ListOptions{All: true})
	if err != nil {
		return PruneResult{}, err
	}
	var res PruneResult
	cutoff := time.Time{}
	if opts.OlderThan > 0 {
		cutoff = time.Now().Add(-opts.OlderThan)
	}
	for _, c := range listed {
		if !shouldPrune(c, opts, cutoff) {
			res.Skipped = append(res.Skipped, c.ID)
			continue
		}
		if err := a.Kill(ctx, c.ID); err != nil {
			res.Skipped = append(res.Skipped, c.ID)
			continue
		}
		res.Pruned = append(res.Pruned, c.ID)
	}
	return res, nil
}

type Colima struct{}

func (Colima) ID() string     { return "colima" }
func (Colima) Binary() string { return "colima" }

func (c Colima) Available(_ context.Context) bool {
	if _, err := exec.LookPath("colima"); err != nil {
		return false
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return true
}

func (c Colima) Version(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "colima", "version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0]), nil
}

func (c Colima) List(ctx context.Context, opts ListOptions) ([]Container, error) {
	return dockerLikeRuntime{binary: "docker"}.List(ctx, opts)
}

func (c Colima) Kill(ctx context.Context, id string) error {
	return dockerLikeRuntime{binary: "docker"}.Kill(ctx, id)
}

func (c Colima) Prune(ctx context.Context, opts PruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{id: "colima", binary: "docker"}.Prune(ctx, opts)
}
