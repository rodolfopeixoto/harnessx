// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	mcppkg "github.com/ropeixoto/harnessx/internal/mcppkg"
)

// registryURL points at the canonical modelcontextprotocol servers
// listing on GitHub. We cache locally for 24h so repeated `harness mcp
// search` calls are offline-friendly.
const (
	registryURL = "https://raw.githubusercontent.com/modelcontextprotocol/servers/main/README.md"
	registryTTL = 24 * time.Hour
)

func mcpSearchCmd() *cobra.Command {
	var (
		refresh bool
	)
	c := &cobra.Command{
		Use:   "search [term]",
		Short: "Search the modelcontextprotocol/servers registry (cached 24h)",
		Long: `Pulls the README.md from github.com/modelcontextprotocol/servers, caches
under .harness/cache/mcp-registry.md, and greps for the term. Bundled
templates (visible via 'harness mcp templates') always appear first.
Pass --refresh to force a fresh fetch.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			term := ""
			if len(args) > 0 {
				term = strings.ToLower(args[0])
			}
			root, err := cwd()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			body, source, err := loadRegistry(root, refresh)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "registry source: %s\n\n", source)

			// bundled hits first
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ORIGIN\tNAME\tHINT")
			bundled, _ := mcppkg.List()
			for _, n := range bundled {
				if term == "" || strings.Contains(n, term) {
					fmt.Fprintf(tw, "bundled\t%s\tharness mcp install %s\n", n, n)
				}
			}
			for _, hit := range searchRegistry(body, term) {
				fmt.Fprintf(tw, "registry\t%s\t%s\n", hit.Name, hit.Link)
			}
			return tw.Flush()
		},
	}
	c.Flags().BoolVar(&refresh, "refresh", false, "ignore the cached registry copy and fetch a fresh one")
	return c
}

func loadRegistry(root string, force bool) ([]byte, string, error) {
	cachePath := filepath.Join(root, ".harness", "cache", "mcp-registry.md")
	if !force {
		if info, err := os.Stat(cachePath); err == nil && time.Since(info.ModTime()) < registryTTL {
			b, err := os.ReadFile(cachePath)
			if err == nil {
				return b, cachePath + " (cached)", nil
			}
		}
	}
	body, err := fetchRegistry(registryURL)
	if err != nil {
		// fall back to stale cache if available
		if b, cerr := os.ReadFile(cachePath); cerr == nil {
			return b, cachePath + " (stale, fetch failed: " + err.Error() + ")", nil
		}
		return nil, "", err
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return body, registryURL + " (no cache: " + err.Error() + ")", nil
	}
	if err := os.WriteFile(cachePath, body, 0o644); err != nil {
		return body, registryURL + " (no cache: " + err.Error() + ")", nil
	}
	return body, registryURL + " (fresh)", nil
}

var fetchRegistry = func(url string) ([]byte, error) {
	cli := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp registry: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

type registryHit struct {
	Name string
	Link string
}

// searchRegistry scrapes the README for `* **[name](link)** description`
// list entries — the format the official registry uses. It is a
// best-effort parser; users wanting structured discovery should run
// the upstream `mcp` CLI instead.
func searchRegistry(body []byte, term string) []registryHit {
	hits := []registryHit{}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") {
			continue
		}
		// pattern: "- **[name](https://...)** - desc"
		start := strings.Index(line, "[")
		mid := strings.Index(line, "](")
		end := strings.Index(line, ")")
		if start < 0 || mid <= start || end <= mid {
			continue
		}
		name := line[start+1 : mid]
		link := line[mid+2 : end]
		if !strings.HasPrefix(link, "https://") {
			continue
		}
		if term != "" && !strings.Contains(strings.ToLower(name+" "+link), term) {
			continue
		}
		hits = append(hits, registryHit{Name: name, Link: link})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Name < hits[j].Name })
	if len(hits) > 30 {
		hits = hits[:30]
	}
	return hits
}

var _ = json.Decoder{} // keep encoding/json referenced for future structured registry support
