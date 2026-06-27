// SPDX-License-Identifier: MIT

package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type RunSpec struct {
	Image      string
	Cmd        []string
	WorkingDir string
	Binds      []BindMount
	Env        map[string]string
	Stdin      string
	AutoRemove bool
	Network    string
	Pull       string
	Timeout    time.Duration
}

type BindMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

type RunResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

type Image struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
}

type ImagePruneOptions struct {
	OlderThan   time.Duration
	IUnderstand bool
}

type RuntimeRunner interface {
	Run(ctx context.Context, spec RunSpec) (RunResult, error)
}

type ImageManager interface {
	ListImages(ctx context.Context) ([]Image, error)
	PruneImages(ctx context.Context, opts ImagePruneOptions) (PruneResult, error)
}

func (d dockerLikeRuntime) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	if spec.Image == "" {
		return RunResult{}, errors.New("containers: RunSpec.Image required")
	}
	args := []string{"run"}
	if spec.AutoRemove {
		args = append(args, "--rm")
	}
	if spec.WorkingDir != "" {
		args = append(args, "-w", spec.WorkingDir)
	}
	if spec.Network != "" {
		args = append(args, "--network", spec.Network)
	}
	if spec.Pull != "" {
		args = append(args, "--pull", spec.Pull)
	}
	for _, b := range spec.Binds {
		mode := "rw"
		if b.ReadOnly {
			mode = "ro"
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s:%s", b.HostPath, b.ContainerPath, mode))
	}
	for k, v := range spec.Env {
		args = append(args, "-e", k+"="+v)
	}
	if spec.Stdin != "" {
		args = append(args, "-i")
	}
	args = append(args, spec.Image)
	args = append(args, spec.Cmd...)

	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(rctx, d.binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if spec.Stdin != "" {
		cmd.Stdin = strings.NewReader(spec.Stdin)
	}
	start := time.Now()
	err := cmd.Run()
	res := RunResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: time.Since(start),
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	if err != nil {
		return res, fmt.Errorf("%s run: %w", d.binary, err)
	}
	return res, nil
}

func (d dockerLikeRuntime) ListImages(ctx context.Context) ([]Image, error) {
	args := []string{"images", "--format", "{{json .}}", "--no-trunc"}
	out, err := exec.CommandContext(ctx, d.binary, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s images: %w", d.binary, err)
	}
	rows := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]Image, 0, len(rows))
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(row), &raw); err != nil {
			continue
		}
		img := Image{
			ID:         firstNonEmpty(raw, "ID"),
			Repository: firstNonEmpty(raw, "Repository"),
			Tag:        firstNonEmpty(raw, "Tag"),
		}
		if v := firstNonEmpty(raw, "CreatedAt", "Created"); v != "" {
			img.CreatedAt = parseFlexibleTime(v)
		}
		res = append(res, img)
	}
	return res, nil
}

func (d dockerLikeRuntime) PruneImages(ctx context.Context, opts ImagePruneOptions) (PruneResult, error) {
	if !opts.IUnderstand {
		return PruneResult{}, errors.New("containers: image prune requires IUnderstand=true or HARNESS_CONTAINERS_I_UNDERSTAND=1")
	}
	args := []string{"image", "prune", "-f"}
	if opts.OlderThan > 0 {
		hours := int(opts.OlderThan.Hours())
		args = append(args, "--filter", fmt.Sprintf("until=%dh", hours))
	}
	out, err := exec.CommandContext(ctx, d.binary, args...).CombinedOutput()
	if err != nil {
		return PruneResult{}, fmt.Errorf("%s image prune: %w: %s", d.binary, err, strings.TrimSpace(string(out)))
	}
	return PruneResult{}, nil
}

func (d Docker) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	return dockerLikeRuntime{id: "docker", binary: "docker"}.Run(ctx, spec)
}
func (d Docker) ListImages(ctx context.Context) ([]Image, error) {
	return dockerLikeRuntime{binary: "docker"}.ListImages(ctx)
}
func (d Docker) PruneImages(ctx context.Context, o ImagePruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{binary: "docker"}.PruneImages(ctx, o)
}

func (p Podman) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	return dockerLikeRuntime{id: "podman", binary: "podman"}.Run(ctx, spec)
}
func (p Podman) ListImages(ctx context.Context) ([]Image, error) {
	return dockerLikeRuntime{binary: "podman"}.ListImages(ctx)
}
func (p Podman) PruneImages(ctx context.Context, o ImagePruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{binary: "podman"}.PruneImages(ctx, o)
}

func (o OrbStack) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	return dockerLikeRuntime{id: "orbstack", binary: "docker"}.Run(ctx, spec)
}
func (o OrbStack) ListImages(ctx context.Context) ([]Image, error) {
	return dockerLikeRuntime{binary: "docker"}.ListImages(ctx)
}
func (o OrbStack) PruneImages(ctx context.Context, opts ImagePruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{binary: "docker"}.PruneImages(ctx, opts)
}

func (c Colima) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	return dockerLikeRuntime{id: "colima", binary: "docker"}.Run(ctx, spec)
}
func (c Colima) ListImages(ctx context.Context) ([]Image, error) {
	return dockerLikeRuntime{binary: "docker"}.ListImages(ctx)
}
func (c Colima) PruneImages(ctx context.Context, opts ImagePruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{binary: "docker"}.PruneImages(ctx, opts)
}

func (a AppleContainer) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	return RunResult{}, errors.New("containers: apple_container Run not yet wired; pick docker/podman via harness runtime set")
}

var appleContainerBinary = "container"

func (a AppleContainer) ListImages(ctx context.Context) ([]Image, error) {
	return appleListImages(ctx, appleContainerBinary)
}

func appleListImages(ctx context.Context, binary string) ([]Image, error) {
	out, err := exec.CommandContext(ctx, binary, "images", "list", "--format", "json").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("apple_container images list: exit %d: %s", exitErr.ExitCode(), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("apple_container images list: %w", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return nil, nil
	}
	if trimmed[0] == '[' {
		return parseAppleImagesArray([]byte(trimmed))
	}
	return parseAppleImagesNDJSON([]byte(trimmed))
}

func parseAppleImagesArray(b []byte) ([]Image, error) {
	var raw []map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("apple_container images parse: %w", err)
	}
	out := make([]Image, 0, len(raw))
	for _, r := range raw {
		out = append(out, appleImageFromMap(r))
	}
	return out, nil
}

func parseAppleImagesNDJSON(b []byte) ([]Image, error) {
	lines := strings.Split(string(b), "\n")
	out := make([]Image, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var r map[string]any
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		out = append(out, appleImageFromMap(r))
	}
	return out, nil
}

func appleImageFromMap(r map[string]any) Image {
	img := Image{
		ID:         firstNonEmpty(r, "id", "ID", "Digest", "digest"),
		Repository: firstNonEmpty(r, "repository", "Repository", "name", "Name"),
		Tag:        firstNonEmpty(r, "tag", "Tag", "reference", "Reference"),
	}
	if s, ok := r["size"].(float64); ok {
		img.SizeBytes = int64(s)
	} else if s, ok := r["Size"].(float64); ok {
		img.SizeBytes = int64(s)
	}
	if v := firstNonEmpty(r, "created", "Created", "createdAt", "CreatedAt"); v != "" {
		img.CreatedAt = parseFlexibleTime(v)
	}
	return img
}

func (a AppleContainer) PruneImages(ctx context.Context, _ ImagePruneOptions) (PruneResult, error) {
	return PruneResult{}, errors.New("containers: apple_container PruneImages not yet wired")
}

var _ = io.EOF
