// SPDX-License-Identifier: MIT

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed hooks/pre-push.sh
var embeddedPrePushHook []byte

//go:embed hooks/pre-commit.sh
var embeddedPreCommitHook []byte

//go:embed hooks/commit-msg.sh
var embeddedCommitMsgHook []byte

const (
	prePushHookMarker = "Managed by HarnessX"
	hookFileMode      = 0o755
)

type hookSpec struct {
	name string
	body []byte
}

func bundledHooks() map[string]hookSpec {
	return map[string]hookSpec{
		"pre-push":   {name: "pre-push", body: embeddedPrePushHook},
		"pre-commit": {name: "pre-commit", body: embeddedPreCommitHook},
		"commit-msg": {name: "commit-msg", body: embeddedCommitMsgHook},
	}
}

func newInstallGitHooksCmd() *cobra.Command {
	var (
		force     bool
		hooksFlag string
	)
	c := &cobra.Command{
		Use:   "install-git-hooks",
		Short: "Install HarnessX-managed git hooks into .git/hooks",
		Long: `Writes embedded git hooks into .git/hooks/. Default installs
pre-push only. Pass --hooks=pre-push,pre-commit,commit-msg to install
others (or --hooks=all). Refuses to overwrite a foreign hook unless
--force is supplied. Bypass at commit/push time with:
  HARNESS_SKIP_CI=1         (pre-push)
  HARNESS_SKIP_PRECOMMIT=1  (pre-commit)
  HARNESS_SKIP_COMMITMSG=1  (commit-msg)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := cwd()
			if err != nil {
				return err
			}
			hooks := parseHooksFlag(hooksFlag)
			installed, err := InstallHooks(dir, hooks, force)
			if err != nil {
				return err
			}
			for _, p := range installed {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ installed %s\n", p)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite an existing non-HarnessX hook")
	c.Flags().StringVar(&hooksFlag, "hooks", "pre-push", "csv of hooks to install (pre-push,pre-commit,commit-msg|all)")
	return c
}

func parseHooksFlag(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "all" {
		keys := []string{}
		for k := range bundledHooks() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// InstallPrePushHook is preserved for backward compatibility with cmd_init
// callers that always want the pre-push hook installed silently.
func InstallPrePushHook(root string, force bool) (string, error) {
	res, err := InstallHooks(root, []string{"pre-push"}, force)
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return filepath.Join(root, ".git", "hooks", "pre-push"), nil
	}
	return res[0], nil
}

func InstallHooks(root string, hooks []string, force bool) ([]string, error) {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("not a git repo: %s", root)
	}
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return nil, fmt.Errorf("create hooks dir: %w", err)
	}
	bundle := bundledHooks()
	var installed []string
	for _, name := range hooks {
		spec, ok := bundle[name]
		if !ok {
			return installed, fmt.Errorf("install-git-hooks: unknown hook %q (have %v)", name, hookNames(bundle))
		}
		dst := filepath.Join(hooksDir, spec.name)
		if existing, err := os.ReadFile(dst); err == nil && !force {
			if !isHarnessManagedHook(existing) {
				return installed, fmt.Errorf("%s exists and is not HarnessX-managed; use --force: %s", spec.name, dst)
			}
		}
		if err := os.WriteFile(dst, spec.body, hookFileMode); err != nil {
			return installed, fmt.Errorf("write %s: %w", spec.name, err)
		}
		installed = append(installed, dst)
	}
	return installed, nil
}

func hookNames(m map[string]hookSpec) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func isHarnessManagedHook(body []byte) bool {
	return strings.Contains(string(body), prePushHookMarker)
}
