// SPDX-License-Identifier: MIT

// Package agentcmd implements the `harness agent …` subcommands.
package agentcmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/adapters/sqlite"
	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/certify"
	"github.com/ropeixoto/harnessx/internal/agents/fake"
	httpadapter "github.com/ropeixoto/harnessx/internal/agents/http"
	yamladapter "github.com/ropeixoto/harnessx/internal/agents/yaml"
	"github.com/ropeixoto/harnessx/internal/domain"
	"github.com/ropeixoto/harnessx/internal/platform/config"
	"github.com/ropeixoto/harnessx/internal/platform/ids"
	"github.com/ropeixoto/harnessx/internal/platform/paths"
)

//go:embed bundled/*.yaml
var bundledFS embed.FS

// LoadAll resolves adapters from (1) project .harness/config/agents/*.yaml and
// (2) bundled templates. Project entries override bundled entries by ID.
// Returns a registry plus the per-id source path for `agent list`.
func LoadAll(root string) (*agents.Registry, map[string]string, error) {
	reg := agents.NewRegistry()
	sources := map[string]string{}

	type discovered struct {
		spec   yamladapter.Spec
		source string
	}
	picks := map[string]discovered{}

	// Bundled first.
	bundled, err := bundledFS.ReadDir("bundled")
	if err == nil {
		for _, e := range bundled {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			b, err := bundledFS.ReadFile("bundled/" + e.Name())
			if err != nil {
				continue
			}
			s, err := parseSpec(b)
			if err != nil {
				return nil, nil, fmt.Errorf("bundled %s: %w", e.Name(), err)
			}
			picks[s.ID] = discovered{spec: s, source: "bundled:" + e.Name()}
		}
	}

	// Project YAMLs override.
	projDir := filepath.Join(root, ".harness", "config", "agents")
	if entries, err := os.ReadDir(projDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			path := filepath.Join(projDir, e.Name())
			s, err := yamladapter.Load(path)
			if err != nil {
				return nil, nil, err
			}
			picks[s.ID] = discovered{spec: s, source: path}
		}
	}

	ids := make([]string, 0, len(picks))
	for id := range picks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		d := picks[id]
		ad := buildAdapter(d.spec)
		if err := reg.Register(ad); err != nil {
			return nil, nil, err
		}
		sources[id] = d.source
	}
	return reg, sources, nil
}

func buildAdapter(s yamladapter.Spec) agents.AgentAdapter {
	if s.Type == "api" {
		return httpadapter.New(s)
	}
	return yamladapter.New(s)
}

func parseSpec(b []byte) (yamladapter.Spec, error) {
	tmp, err := os.CreateTemp("", "harnessx-spec-*.yaml")
	if err != nil {
		return yamladapter.Spec{}, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return yamladapter.Spec{}, err
	}
	_ = tmp.Close()
	return yamladapter.Load(tmp.Name())
}

// --- list -------------------------------------------------------------------

func List(out io.Writer, startDir string) error {
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	reg, sources, err := LoadAll(root)
	if err != nil {
		return err
	}
	if len(reg.IDs()) == 0 {
		fmt.Fprintln(out, "no agent adapters registered")
		return nil
	}
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	var repo *sqlite.Repo
	if _, err := os.Stat(config.Resolve(root, cfg.Database.Path)); err == nil {
		repo, _ = sqlite.Open(config.Resolve(root, cfg.Database.Path))
		if repo != nil {
			defer repo.Close()
		}
	}
	fmt.Fprintf(out, "%-12s %-22s %-10s %s\n", "ID", "NAME", "CERT", "SOURCE")
	for _, id := range reg.IDs() {
		a, _ := reg.Get(id)
		cert := "—"
		if repo != nil {
			if c, err := repo.LatestAgentCertification(context.Background(), id); err == nil {
				cert = fmt.Sprintf("%s/%d", c.Status, c.Score)
			}
		}
		fmt.Fprintf(out, "%-12s %-22s %-10s %s\n", a.ID(), truncate(a.Name(), 22), cert, sources[id])
	}
	return nil
}

// --- add --------------------------------------------------------------------

func Add(out io.Writer, startDir, id string) error {
	if id == "" {
		return errors.New("agent add: missing id")
	}
	root, err := paths.FindProjectRoot(startDir)
	if err != nil {
		return err
	}
	dst := filepath.Join(root, ".harness", "config", "agents")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	dstFile := filepath.Join(dst, id+".yaml")
	if _, err := os.Stat(dstFile); err == nil {
		return fmt.Errorf("agent add: %s already exists", dstFile)
	}
	b, err := bundledFS.ReadFile("bundled/" + id + ".yaml")
	if err != nil {
		return fmt.Errorf("agent add: no bundled adapter for %q (available: %s)", id, strings.Join(listBundled(), ", "))
	}
	if err := os.WriteFile(dstFile, b, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", dstFile)
	return nil
}

func listBundled() []string {
	entries, _ := bundledFS.ReadDir("bundled")
	var out []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yaml") {
			out = append(out, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	sort.Strings(out)
	return out
}

// --- discover ---------------------------------------------------------------

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

// --- certify ----------------------------------------------------------------

type CertifyOptions struct {
	ID            string
	SkipRun       bool
	SimpleTimeout time.Duration
	StartDir      string
	Override      agents.AgentAdapter
}

func Certify(ctx context.Context, out io.Writer, opts CertifyOptions) (certify.Result, error) {
	root, err := paths.FindProjectRoot(opts.StartDir)
	if err != nil {
		return certify.Result{}, err
	}
	var a agents.AgentAdapter
	if opts.Override != nil {
		a = opts.Override
	} else {
		reg, _, err := LoadAll(root)
		if err != nil {
			return certify.Result{}, err
		}
		var ok bool
		a, ok = reg.Get(opts.ID)
		if !ok {
			return certify.Result{}, fmt.Errorf("agent certify: %q not registered", opts.ID)
		}
	}

	res := certify.Run(ctx, a, certify.Options{SkipRun: opts.SkipRun, SimpleTimeout: opts.SimpleTimeout})

	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err == nil {
		repo, err := sqlite.Open(dbPath)
		if err == nil {
			defer repo.Close()
			_ = repo.WriteAgentCertification(ctx, domain.AgentCertification{
				ID: ids.New(), AgentID: a.ID(), CLIVersion: res.CLIVersion,
				AdapterVersion: "1", Score: res.Score, Status: res.Status,
				DetailsJSON: res.DetailsJSON(), CertifiedAt: time.Now().UTC(),
			})
		}
	}

	renderCertification(out, a, res)
	return res, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
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

// Ensure fake adapter package stays referenced when no caller imports it
// (the test e2e exercises it indirectly).
var _ fs.FS = bundledFS
var _ = fake.New
