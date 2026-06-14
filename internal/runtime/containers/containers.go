// SPDX-License-Identifier: MIT

package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Item struct {
	ID        string
	Name      string
	Image     string
	Status    string
	SizeBytes int64
	CreatedAt time.Time
}

type Lister interface {
	List(ctx context.Context) ([]Item, error)
}

type RealLister struct {
	Binary string
}

func (r RealLister) binary() string {
	if r.Binary != "" {
		return r.Binary
	}
	return constants.DefaultDockerBinary
}

func (r RealLister) List(ctx context.Context) ([]Item, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.DefaultDockerStatsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.binary(), "ps", "-a", "--format", "{{json .}}")
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("containers: docker ps: %w (%s)", err, strings.TrimSpace(errBuf.String()))
	}
	return parseDockerPS(out.String())
}

type psRow struct {
	ID        string `json:"ID"`
	Names     string `json:"Names"`
	Image     string `json:"Image"`
	Status    string `json:"Status"`
	CreatedAt string `json:"CreatedAt"`
	Size      string `json:"Size"`
}

func parseDockerPS(stdout string) ([]Item, error) {
	var items []Item
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var row psRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		items = append(items, Item{
			ID:        row.ID,
			Name:      row.Names,
			Image:     row.Image,
			Status:    row.Status,
			CreatedAt: parseDockerTime(row.CreatedAt),
		})
	}
	return items, nil
}

func parseDockerTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05 -0700 MST"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
