// SPDX-License-Identifier: MIT

package critic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/router"
)

type Request struct {
	Diff           string
	OriginalPrompt string
	AdapterID      string
}

type Verdict struct {
	Score       float64
	AdapterID   string
	Concerns    []string
	Suggestions []string
	Raw         string
}

func PickCritic(reg *agents.Registry) (string, bool) {
	for _, tags := range [][]string{{"review", "critic"}, {"review"}, {"reasoning"}, {"code"}} {
		if c, ok := router.Pick(tags, reg); ok {
			return c.AdapterID, true
		}
	}
	for _, id := range reg.IDs() {
		return id, true
	}
	return "", false
}

func Critique(ctx context.Context, req Request, reg *agents.Registry) (Verdict, error) {
	if reg == nil {
		return Verdict{}, errors.New("critic: nil registry")
	}
	id := req.AdapterID
	if id == "" {
		picked, ok := PickCritic(reg)
		if !ok {
			return Verdict{}, errors.New("critic: no adapter available")
		}
		id = picked
	}
	a, ok := reg.Get(id)
	if !ok {
		return Verdict{}, errors.New("critic: adapter not registered: " + id)
	}
	prompt := buildPrompt(req)
	res := a.Run(ctx, agents.AgentRequest{Prompt: prompt, Timeout: 60 * time.Second})
	if res.Err != nil {
		return Verdict{AdapterID: id}, res.Err
	}
	return parseVerdict(id, string(res.Output.Stdout)+res.Output.FinalMessage), nil
}

func buildPrompt(req Request) string {
	var b strings.Builder
	b.WriteString("Review the following diff for the request.\n\n")
	b.WriteString("# Request\n")
	b.WriteString(req.OriginalPrompt)
	b.WriteString("\n\n# Diff\n")
	b.WriteString(req.Diff)
	b.WriteString("\n\nReply with one line `score: N/10` plus bullet lines under `concerns:` and `suggestions:`.")
	return b.String()
}

func parseVerdict(adapterID, raw string) Verdict {
	v := Verdict{AdapterID: adapterID, Raw: raw}
	var section string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(strings.ToLower(line), "score:"):
			v.Score = parseScore(line)
		case strings.HasPrefix(strings.ToLower(line), "concerns:"):
			section = "concerns"
		case strings.HasPrefix(strings.ToLower(line), "suggestions:"):
			section = "suggestions"
		case strings.HasPrefix(line, "-"), strings.HasPrefix(line, "*"):
			item := strings.TrimSpace(strings.TrimLeft(line, "-* "))
			switch section {
			case "concerns":
				v.Concerns = append(v.Concerns, item)
			case "suggestions":
				v.Suggestions = append(v.Suggestions, item)
			}
		}
	}
	return v
}

func parseScore(line string) float64 {
	idx := strings.Index(line, ":")
	if idx < 0 || idx == len(line)-1 {
		return 0
	}
	tail := strings.TrimSpace(line[idx+1:])
	if slash := strings.Index(tail, "/"); slash > 0 {
		tail = tail[:slash]
	}
	var n float64
	for _, r := range tail {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + float64(r-'0')
	}
	return n
}
