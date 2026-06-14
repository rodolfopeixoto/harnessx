package yaml

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ropeixoto/harnessx/internal/agents"
)

type stubRunner struct {
	stdout, stderr []byte
	exit           int
	err            error
	gotArgs        []string
	gotStdin       string
}

func (r *stubRunner) Run(ctx context.Context, name string, args []string, stdin, wd string) ([]byte, []byte, int, error) {
	r.gotArgs = args
	r.gotStdin = stdin
	return r.stdout, r.stderr, r.exit, r.err
}

func TestLoad_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake.yaml")
	require.NoError(t, writeFile(path, `id: fake
name: Fake
enabled: true
type: cli
command:
  binary: fake-cli
  check: fake-cli --version
capabilities:
  text: true
  max_context_tokens: 100000
strengths: [implementation]
models:
  default: m1
execution:
  prompt_mode: stdin
  timeout_seconds: 60
run:
  args: ["run", "--model", "{{model}}", "--json"]
output:
  format: jsonl
  final_message_json_path: "$.message"
  usage_json_path: "$.usage"
failure_detection:
  rate_limit: ["rate limit", "quota exceeded"]
  auth: ["unauthorized"]
cost:
  mode: estimated
  input_token_price_per_1m: 1.0
  output_token_price_per_1m: 2.0
`))
	s, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "fake", s.ID)
	require.Equal(t, "m1", s.Models["default"])
	require.Equal(t, "jsonl", s.Output.Format)
}

func TestRun_SuccessAndUsageParsing(t *testing.T) {
	r := &stubRunner{
		stdout: []byte(`{"message":"hello","usage":{"input_tokens":10,"output_tokens":3}}` + "\n"),
		exit:   0,
	}
	a := New(Spec{
		ID: "x", Name: "X", Type: "cli",
	})
	a.Spec.Command.Binary = "x-cli"
	a.Spec.Run.Args = []string{"run", "--model", "{{model}}"}
	a.Spec.Output.Format = "jsonl"
	a.Spec.Output.FinalMessageJSONPath = "$.message"
	a.Spec.Output.UsageJSONPath = "$.usage"
	a.Spec.Cost.InputTokenPricePer1M = 1.0
	a.Spec.Cost.OutputTokenPricePer1M = 2.0
	a.Lookup = func(string) (string, error) { return "/usr/bin/x-cli", nil }
	a.Runner = r

	res := a.Run(context.Background(), agents.AgentRequest{Prompt: "hi", Model: "m1"})
	require.NoError(t, res.Err)
	require.Equal(t, 0, res.ExitCode)
	require.Equal(t, "hello", res.Output.FinalMessage)
	require.Equal(t, 10, res.Usage.InputTokens)
	require.Equal(t, 3, res.Usage.OutputTokens)
	require.Equal(t, "reported", res.Usage.Mode)
	require.Equal(t, agents.FailureNone, res.Failure)
	require.Equal(t, []string{"run", "--model", "m1"}, r.gotArgs)
	require.Equal(t, "hi", r.gotStdin)
}

func TestRun_ClassifyRateLimit(t *testing.T) {
	r := &stubRunner{stderr: []byte("HTTP 429: rate limit exceeded"), exit: 1, err: errors.New("exit 1")}
	a := New(Spec{ID: "x", Name: "X", Type: "cli"})
	a.Spec.Command.Binary = "x"
	a.Spec.FailureDetection = map[string][]string{"rate_limit": {"rate limit"}}
	a.Lookup = func(string) (string, error) { return "/x", nil }
	a.Runner = r

	res := a.Run(context.Background(), agents.AgentRequest{Prompt: "p"})
	require.Equal(t, agents.FailureRateLimit, res.Failure)
	require.True(t, res.Failure.IsRecoverable())
}

func TestExtractJSONPath_NestedAndJSONL(t *testing.T) {
	got := extractJSONPath([]byte(`{"a":{"b":"c"}}`), "json", "$.a.b")
	require.Equal(t, "c", got)
	got = extractJSONPath([]byte("{\"x\":1}\n{\"message\":\"final\"}\n"), "jsonl", "$.message")
	require.Equal(t, "final", got)
}

func writeFile(path, body string) error {
	return writeFileImpl(path, body)
}
