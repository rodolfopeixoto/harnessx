package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/adapters/execprobe"
	"github.com/ropeixoto/harnessx/internal/app/doctor"
)

func TestRenderDoctor_PlainOutputContainsKeySections(t *testing.T) {
	SetPlain(true)

	r := doctor.Report{
		OS: "darwin", Arch: "amd64",
		Tools: []doctor.Entry{
			{Spec: doctor.ProbeSpec{Binary: "git", Label: "Git", Required: true, Category: "tool"},
				Result: execprobe.Result{Binary: "git", Present: true, Version: "git version 2.44.0"}},
			{Spec: doctor.ProbeSpec{Binary: "docker", Label: "Docker", Required: false, Category: "tool"},
				Result: execprobe.Result{Binary: "docker", Present: false}},
		},
		Agents: []doctor.Entry{
			{Spec: doctor.ProbeSpec{Binary: "claude", Label: "Claude Code", Required: false, Category: "agent"},
				Result: execprobe.Result{Binary: "claude", Present: true, Version: "claude 1.0.0"}},
		},
		Project: doctor.ProjectInfo{Root: "/tmp/p", HarnessReady: true},
	}
	var buf bytes.Buffer
	RenderDoctor(&buf, r)
	out := buf.String()
	for _, want := range []string{"HarnessX Doctor", "Tools", "Agents", "Project", "Git", "Docker", "Claude Code", "/tmp/p"} {
		require.Truef(t, strings.Contains(out, want), "want %q in:\n%s", want, out)
	}
}
