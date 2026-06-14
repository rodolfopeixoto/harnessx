// SPDX-License-Identifier: MIT

package context

import (
	"context"
	"strings"

	"github.com/ropeixoto/harnessx/internal/index"
)

type TestMapProvider struct{}

func (TestMapProvider) Name() string { return "test_map" }

func (TestMapProvider) Apply(ctx context.Context, root string, pack *Pack) error {
	var tm index.TestMap
	if err := index.ReadMap(root, index.MapTests, &tm); err != nil {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	// Add tests that share a substring with any relevant file path. Cheap
	// heuristic — Phase 5 doesn't ship a real source-to-test resolver.
	want := map[string]bool{}
	for _, f := range pack.RelevantFiles {
		base := lastSegment(f.Path)
		want[base] = true
	}
	for _, suite := range tm.Suites {
		for _, t := range suite.Files {
			tbase := lastSegment(t)
			for w := range want {
				if w != "" && strings.Contains(tbase, strings.TrimSuffix(w, ".go")) {
					pack.RelatedTests = appendUnique(pack.RelatedTests, t)
					break
				}
			}
		}
	}
	pack.Stats.ProvidersRan++
	return nil
}

func lastSegment(p string) string {
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func appendUnique(in []string, s string) []string {
	for _, v := range in {
		if v == s {
			return in
		}
	}
	return append(in, s)
}
