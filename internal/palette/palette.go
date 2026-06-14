// SPDX-License-Identifier: MIT

package palette

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type Hit struct {
	Source     string `json:"source"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`
	RouterPath string `json:"router_path,omitempty"`
	Score      int    `json:"score"`
}

type Source interface {
	Name() string
	Search(ctx context.Context, q string) ([]Hit, error)
}

type Palette struct {
	mu      sync.RWMutex
	sources []Source
	limit   int
}

const DefaultLimit = 25

func New(sources ...Source) *Palette {
	return &Palette{sources: sources, limit: DefaultLimit}
}

func (p *Palette) WithLimit(n int) *Palette {
	p.limit = n
	return p
}

func (p *Palette) Search(ctx context.Context, q string) ([]Hit, error) {
	if strings.TrimSpace(q) == "" {
		return nil, nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	var all []Hit
	for _, s := range p.sources {
		hits, err := s.Search(ctx, q)
		if err != nil {
			return nil, err
		}
		all = append(all, hits...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Score != all[j].Score {
			return all[i].Score > all[j].Score
		}
		return all[i].Title < all[j].Title
	})
	if p.limit > 0 && len(all) > p.limit {
		all = all[:p.limit]
	}
	return all, nil
}

func Score(needle, haystack string) int {
	if needle == "" || haystack == "" {
		return 0
	}
	needle = strings.ToLower(needle)
	haystack = strings.ToLower(haystack)
	if haystack == needle {
		return 100
	}
	if strings.HasPrefix(haystack, needle) {
		return 80
	}
	if strings.Contains(haystack, needle) {
		return 60
	}
	score := 0
	hi := 0
	for _, r := range needle {
		idx := strings.IndexRune(haystack[hi:], r)
		if idx < 0 {
			return 0
		}
		hi += idx + 1
		score++
	}
	return 20 + score
}
