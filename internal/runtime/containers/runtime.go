// SPDX-License-Identifier: MIT

package containers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Runtime interface {
	ID() string
	Binary() string
	Available(ctx context.Context) bool
	Version(ctx context.Context) (string, error)
	List(ctx context.Context, opts ListOptions) ([]Container, error)
	Kill(ctx context.Context, id string) error
	Prune(ctx context.Context, opts PruneOptions) (PruneResult, error)
	Run(ctx context.Context, spec RunSpec) (RunResult, error)
	ListImages(ctx context.Context) ([]Image, error)
	PruneImages(ctx context.Context, opts ImagePruneOptions) (PruneResult, error)
}

type ListOptions struct {
	All bool
}

type Container struct {
	ID        string
	Name      string
	Image     string
	Status    string
	State     string
	CreatedAt time.Time
}

type PruneOptions struct {
	Stopped     bool
	All         bool
	OlderThan   time.Duration
	IUnderstand bool
}

type PruneResult struct {
	Pruned     []string
	Skipped    []string
	BytesFreed int64
}

func KnownRuntimeIDs() []string {
	return []string{"apple_container", "docker", "orbstack", "podman", "colima"}
}

func Detect(ctx context.Context) []Runtime {
	all := buildAll()
	var out []Runtime
	for _, r := range all {
		if r.Available(ctx) {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return preferenceRank(out[i].ID()) < preferenceRank(out[j].ID())
	})
	return out
}

func DetectIncluding(ctx context.Context, includeUnavailable bool) []Runtime {
	if !includeUnavailable {
		return Detect(ctx)
	}
	all := buildAll()
	sort.SliceStable(all, func(i, j int) bool {
		return preferenceRank(all[i].ID()) < preferenceRank(all[j].ID())
	})
	return all
}

func ByID(id string) (Runtime, error) {
	for _, r := range buildAll() {
		if r.ID() == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("containers: unknown runtime %q (known: %v)", id, KnownRuntimeIDs())
}

func buildAll() []Runtime {
	return []Runtime{
		&AppleContainer{},
		&Docker{},
		&OrbStack{},
		&Podman{},
		&Colima{},
	}
}

func preferenceRank(id string) int {
	pref := platformPreference()
	for i, v := range pref {
		if v == id {
			return i
		}
	}
	return len(pref) + 100
}

func platformPreference() []string {
	if runtime.GOOS == "darwin" {
		return []string{"apple_container", "docker", "orbstack", "podman", "colima"}
	}
	return []string{"docker", "podman", "orbstack", "colima"}
}

type dockerLikeRuntime struct {
	id     string
	binary string
}

func (d dockerLikeRuntime) ID() string     { return d.id }
func (d dockerLikeRuntime) Binary() string { return d.binary }

func (d dockerLikeRuntime) Available(ctx context.Context) bool {
	_, err := exec.LookPath(d.binary)
	return err == nil
}

func (d dockerLikeRuntime) Version(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, d.binary, "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (d dockerLikeRuntime) List(ctx context.Context, opts ListOptions) ([]Container, error) {
	args := []string{"ps", "--format", "{{json .}}"}
	if opts.All {
		args = append(args, "--all")
	}
	out, err := exec.CommandContext(ctx, d.binary, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s ps: %w", d.binary, err)
	}
	return parseDockerJSON(out)
}

func (d dockerLikeRuntime) Kill(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("containers: empty id")
	}
	if out, err := exec.CommandContext(ctx, d.binary, "rm", "-f", id).CombinedOutput(); err != nil {
		return fmt.Errorf("%s rm -f %s: %w: %s", d.binary, id, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (d dockerLikeRuntime) Prune(ctx context.Context, opts PruneOptions) (PruneResult, error) {
	if !opts.IUnderstand {
		return PruneResult{}, errors.New("containers: prune requires IUnderstand=true or HARNESS_CONTAINERS_I_UNDERSTAND=1")
	}
	listed, err := d.List(ctx, ListOptions{All: true})
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
		if err := d.Kill(ctx, c.ID); err != nil {
			res.Skipped = append(res.Skipped, c.ID)
			continue
		}
		res.Pruned = append(res.Pruned, c.ID)
	}
	return res, nil
}

func shouldPrune(c Container, opts PruneOptions, cutoff time.Time) bool {
	if opts.All {
		if !cutoff.IsZero() && c.CreatedAt.After(cutoff) {
			return false
		}
		return true
	}
	state := strings.ToLower(c.State)
	status := strings.ToLower(c.Status)
	stopped := state == "exited" || strings.Contains(status, "exited") || strings.Contains(status, "dead")
	if opts.Stopped && !stopped {
		return false
	}
	if !cutoff.IsZero() && c.CreatedAt.After(cutoff) {
		return false
	}
	return true
}

func parseDockerJSON(out []byte) ([]Container, error) {
	rows := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]Container, 0, len(rows))
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(row), &raw); err != nil {
			continue
		}
		c := Container{
			ID:     firstNonEmpty(raw, "ID", "Id", "Container"),
			Name:   firstNonEmpty(raw, "Names", "Name"),
			Image:  firstNonEmpty(raw, "Image"),
			Status: firstNonEmpty(raw, "Status"),
			State:  firstNonEmpty(raw, "State"),
		}
		if v := firstNonEmpty(raw, "CreatedAt", "Created"); v != "" {
			c.CreatedAt = parseFlexibleTime(v)
		}
		res = append(res, c)
	}
	return res, nil
}

func firstNonEmpty(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func parseFlexibleTime(s string) time.Time {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

type Docker struct{}

func (Docker) ID() string     { return "docker" }
func (Docker) Binary() string { return "docker" }
func (d Docker) Available(ctx context.Context) bool {
	return dockerLikeRuntime{binary: "docker"}.Available(ctx)
}
func (d Docker) Version(ctx context.Context) (string, error) {
	return dockerLikeRuntime{binary: "docker"}.Version(ctx)
}
func (d Docker) List(ctx context.Context, o ListOptions) ([]Container, error) {
	return dockerLikeRuntime{binary: "docker"}.List(ctx, o)
}
func (d Docker) Kill(ctx context.Context, id string) error {
	return dockerLikeRuntime{binary: "docker"}.Kill(ctx, id)
}
func (d Docker) Prune(ctx context.Context, o PruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{id: "docker", binary: "docker"}.Prune(ctx, o)
}

type Podman struct{}

func (Podman) ID() string     { return "podman" }
func (Podman) Binary() string { return "podman" }
func (p Podman) Available(ctx context.Context) bool {
	return dockerLikeRuntime{binary: "podman"}.Available(ctx)
}
func (p Podman) Version(ctx context.Context) (string, error) {
	return dockerLikeRuntime{binary: "podman"}.Version(ctx)
}
func (p Podman) List(ctx context.Context, o ListOptions) ([]Container, error) {
	return dockerLikeRuntime{binary: "podman"}.List(ctx, o)
}
func (p Podman) Kill(ctx context.Context, id string) error {
	return dockerLikeRuntime{binary: "podman"}.Kill(ctx, id)
}
func (p Podman) Prune(ctx context.Context, o PruneOptions) (PruneResult, error) {
	return dockerLikeRuntime{id: "podman", binary: "podman"}.Prune(ctx, o)
}

type OrbStack struct{}

func (OrbStack) ID() string     { return "orbstack" }
func (OrbStack) Binary() string { return "orbctl" }

func (o OrbStack) Available(ctx context.Context) bool {
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
	return strings.Contains(strings.ToLower(string(out)), "container")
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

func (c Colima) Available(ctx context.Context) bool {
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
