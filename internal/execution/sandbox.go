// SPDX-License-Identifier: MIT

package execution

import (
	"context"
	"fmt"
	"strings"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/runtime/containers"
)

type SandboxMode string

const (
	SandboxHost      SandboxMode = "host"
	SandboxContainer SandboxMode = "container"
)

type SandboxSpec struct {
	Mode    SandboxMode
	Image   string
	EnvKeep []string
}

func runInContainer(ctx context.Context, projectRoot string, sb SandboxSpec, wt Worktree, agentReq agents.AgentRequest, binary string) (agents.AgentResult, error) {
	rt, _, err := containers.Resolve(ctx, projectRoot)
	if err != nil {
		return agents.AgentResult{}, fmt.Errorf("sandbox: resolve runtime: %w", err)
	}
	image := sb.Image
	if image == "" {
		image = "alpine:3.20"
	}
	cmd := append([]string{binary}, agentReq.ExtraArgs...)
	spec := containers.RunSpec{
		Image:      image,
		Cmd:        cmd,
		WorkingDir: "/work",
		Binds:      []containers.BindMount{{HostPath: wt.Path, ContainerPath: "/work", ReadOnly: false}},
		Env:        passthroughEnv(sb.EnvKeep),
		Stdin:      agentReq.Prompt,
		AutoRemove: true,
		Timeout:    agentReq.Timeout,
	}
	res, err := rt.Run(ctx, spec)
	if err != nil {
		return agents.AgentResult{}, err
	}
	out := agents.AgentOutput{Stdout: res.Stdout, Stderr: res.Stderr}
	out.FinalMessage = strings.TrimSpace(string(res.Stdout))
	return agents.AgentResult{
		Output:   out,
		ExitCode: res.ExitCode,
		Duration: res.Duration,
	}, nil
}

func passthroughEnv(keep []string) map[string]string {
	out := map[string]string{}
	for _, k := range keep {
		if v := osLookup(k); v != "" {
			out[k] = v
		}
	}
	return out
}

func osLookup(key string) string {
	return ""
}
