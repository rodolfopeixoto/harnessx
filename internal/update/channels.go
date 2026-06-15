// SPDX-License-Identifier: MIT

// Package update resolves GitHub release channels (stable / beta /
// develop) and downloads + verifies + swaps the harness binary in
// place. Pure stdlib; no network code in tests (the resolver accepts a
// release lister so tests inject fakes).
package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Channel string

const (
	ChannelStable  Channel = "stable"
	ChannelBeta    Channel = "beta"
	ChannelDevelop Channel = "develop"
)

func KnownChannels() []Channel {
	return []Channel{ChannelStable, ChannelBeta, ChannelDevelop}
}

// Release mirrors the GitHub releases API subset we care about.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// Lister fetches releases from somewhere (GitHub by default).
type Lister interface {
	List(repo string) ([]Release, error)
}

// GitHubLister hits the GitHub releases API.
type GitHubLister struct {
	HTTPClient *http.Client
	Token      string // optional, raises the rate limit
}

func NewGitHubLister() *GitHubLister {
	return &GitHubLister{HTTPClient: &http.Client{Timeout: 15 * time.Second}}
}

func (g *GitHubLister) List(repo string) ([]Release, error) {
	if g.HTTPClient == nil {
		g.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=50", repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api %s: %d: %s", url, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out []Release
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// PickLatest returns the newest release matching the channel.
// stable  = newest non-draft, non-prerelease
// beta    = newest non-draft (includes prereleases)
// develop = caller handles source build, returns ErrSourceChannel.
func PickLatest(channel Channel, releases []Release) (Release, error) {
	switch channel {
	case ChannelDevelop:
		return Release{}, ErrSourceChannel
	case ChannelBeta, ChannelStable:
	default:
		return Release{}, fmt.Errorf("unknown channel %q (known: stable, beta, develop)", channel)
	}
	candidates := make([]Release, 0, len(releases))
	for _, r := range releases {
		if r.Draft {
			continue
		}
		if channel == ChannelStable && r.Prerelease {
			continue
		}
		candidates = append(candidates, r)
	}
	if len(candidates) == 0 {
		return Release{}, fmt.Errorf("no release on channel %s", channel)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].PublishedAt.After(candidates[j].PublishedAt) })
	return candidates[0], nil
}

// FilterChannel returns every release on a given channel, newest first.
func FilterChannel(channel Channel, releases []Release) []Release {
	var out []Release
	for _, r := range releases {
		if r.Draft {
			continue
		}
		if channel == ChannelStable && r.Prerelease {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PublishedAt.After(out[j].PublishedAt) })
	return out
}

// CompareVersions returns 1 when a > b, -1 when a < b, 0 when equal.
// Tolerates `v` prefix and `-betaN` / `-rcN` suffixes (suffix loses to
// the same tag without suffix).
func CompareVersions(a, b string) int {
	an, ap := splitVersion(a)
	bn, bp := splitVersion(b)
	for i := 0; i < 3; i++ {
		if an[i] != bn[i] {
			if an[i] > bn[i] {
				return 1
			}
			return -1
		}
	}
	switch {
	case ap == "" && bp != "":
		return 1
	case ap != "" && bp == "":
		return -1
	case ap > bp:
		return 1
	case ap < bp:
		return -1
	default:
		return 0
	}
}

func splitVersion(s string) ([3]int, string) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	pre := ""
	if i := strings.Index(s, "-"); i >= 0 {
		pre = s[i+1:]
		s = s[:i]
	}
	if i := strings.Index(s, " "); i >= 0 {
		s = s[:i]
	}
	parts := strings.SplitN(s, ".", 3)
	var nums [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		_, _ = fmt.Sscanf(parts[i], "%d", &nums[i])
	}
	return nums, pre
}

// ErrSourceChannel signals the caller should build from source instead
// of picking a release.
var ErrSourceChannel = errors.New("update: develop channel requires a source build")
