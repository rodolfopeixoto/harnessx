// SPDX-License-Identifier: MIT

package interactive

import (
	"context"
	"fmt"
	"runtime"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/yaml"
)

type Adapter struct {
	Spec     yaml.Spec
	Strategy Strategy
}

func New(spec yaml.Spec) *Adapter {
	return &Adapter{Spec: spec, Strategy: pickStrategy(spec.Interactive.Strategy)}
}

func pickStrategy(name string) Strategy {
	switch name {
	case "tmux":
		return TmuxStrategy{}
	case "iterm2":
		return ITermStrategy{}
	default:
		return PTYStrategy{}
	}
}

func (a *Adapter) IsExperimental() bool              { return a.Spec.Experimental }
func (a *Adapter) ID() string                        { return a.Spec.ID }
func (a *Adapter) Name() string                      { return a.Spec.Name }
func (a *Adapter) Capabilities() agents.Capabilities { return a.Spec.Capabilities }

func (a *Adapter) Healthcheck(ctx context.Context) agents.HealthcheckResult {
	bin := a.Spec.Interactive.Binary
	if bin == "" {
		return agents.HealthcheckResult{OK: false, Err: "interactive.binary missing in spec"}
	}
	if runtime.GOOS == "windows" {
		return agents.HealthcheckResult{OK: false, Err: "interactive adapter not yet supported on windows"}
	}
	return agents.HealthcheckResult{OK: true, Detail: fmt.Sprintf("strategy=%s binary=%s", a.Strategy.ID(), bin)}
}

func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
	cfg := Config{
		Binary:             a.Spec.Interactive.Binary,
		Args:               a.Spec.Interactive.Args,
		IdleMs:             a.Spec.Interactive.IdleMs,
		HardTimeoutSeconds: a.Spec.Interactive.HardTimeoutSeconds,
		BannerPattern:      a.Spec.Interactive.BannerPattern,
		TmuxSessionName:    a.Spec.Interactive.Tmux.SessionName,
		ITermProfile:       a.Spec.Interactive.ITerm2.Profile,
	}
	res, err := a.Strategy.Run(ctx, cfg, req)
	if err != nil {
		res.Err = err
		if res.Failure == agents.FailureNone {
			res.Failure = agents.FailureTransient
		}
	}
	res.Usage = a.ParseUsage(res.Output)
	return res
}

func (a *Adapter) ParseUsage(out agents.AgentOutput) agents.Usage {
	return agents.Usage{
		InputTokens:  estimateTokens(string(out.Stdout)),
		OutputTokens: estimateTokens(out.FinalMessage),
		Mode:         "estimated",
	}
}

func (a *Adapter) ClassifyFailure(_ agents.AgentOutput, _ int, runErr error) agents.FailureType {
	if runErr == nil {
		return agents.FailureNone
	}
	return agents.FailureTransient
}

func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	return len(s) / 4
}
