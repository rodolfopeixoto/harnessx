// SPDX-License-Identifier: MIT

package agentcmd

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Discover prints a YAML scaffold for wiring a new CLI-based adapter.
func Discover(out io.Writer, binary string) error {
	if binary == "" {
		return errors.New("agent discover: missing binary")
	}
	id := strings.ReplaceAll(strings.TrimSuffix(binary, "-cli"), "/", "-")
	scaffold := fmt.Sprintf(`id: %s
name: %s
enabled: true
type: cli

command:
  binary: %s
  check: %s --version

capabilities:
  text: true
  files: true
  json_output: false
  max_context_tokens: 128000

strengths: []

models:
  default: ""

execution:
  prompt_mode: stdin
  working_directory: project
  timeout_seconds: 600

run:
  args: []

output:
  format: text

failure_detection:
  rate_limit: ["rate limit"]
  auth: ["unauthorized"]

cost:
  mode: estimated
  input_token_price_per_1m: 0.0
  output_token_price_per_1m: 0.0
`, id, titleCase(id), binary, binary)
	fmt.Fprintln(out, "# save the following YAML under .harness/config/agents/"+id+".yaml")
	fmt.Fprintln(out, scaffold)
	return nil
}

// titleCase upper-cases the first rune; used for human-readable adapter
// names in the discover scaffold. strings.Title is deprecated; we don't
// need its Unicode word-boundary logic for ASCII adapter ids.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 32
	}
	return string(r)
}
