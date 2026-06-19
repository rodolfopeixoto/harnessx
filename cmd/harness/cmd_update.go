// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/update"
	"github.com/ropeixoto/harnessx/internal/version"
)

const defaultRepo = "rodolfopeixoto/harnessx"

func newUpdateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade", "self-update"},
		Short:   "Self-update to the latest harness release",
		Long: `Self-update fetches the latest release from GitHub, verifies its
SHA-256, and replaces the running binary in place. Channels:

  stable   — newest non-prerelease tag (default)
  beta     — newest tag including pre-releases (vX.Y.Z-beta*, -rc*)
  develop  — clone the develop branch and build from source (needs git + go)

Examples:
  harness update                       # latest stable
  harness update --channel beta        # opt into pre-releases
  harness update --tag v0.4.0          # pin a specific tag
  harness update --dry-run             # see the plan without swapping
  harness update status                # is there something newer?
  harness update channels --json       # machine-readable channel listing`,
		RunE: runDoUpdate,
	}
	addUpdateFlags(c)
	c.AddCommand(newUpdateStatusCmd(), newUpdateChannelsCmd())
	return c
}

var updateFlags struct {
	repo    string
	tag     string
	channel string
	dryRun  bool
	jsonOut bool
	force   bool
}

func addUpdateFlags(c *cobra.Command) {
	c.Flags().StringVar(&updateFlags.repo, "repo", defaultRepo, "GitHub repo (owner/name)")
	c.Flags().StringVar(&updateFlags.tag, "tag", "", "pin a specific tag (overrides --channel)")
	c.Flags().StringVar(&updateFlags.channel, "channel", "stable", "stable|beta|develop")
	c.Flags().BoolVar(&updateFlags.dryRun, "dry-run", false, "print plan without replacing binary")
	c.Flags().BoolVar(&updateFlags.force, "force", false, "allow downgrade OR reinstall the same tag")
}

func normaliseTag(t string) string {
	if t == "" || t[0] == 'v' {
		return t
	}
	return "v" + t
}

func newUpdateStatusCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "status",
		Short: "Show whether a newer release is available",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lister := update.NewGitHubLister()
			rs, err := lister.List(updateFlags.repo)
			if err != nil {
				return err
			}
			channel := update.Channel(updateFlags.channel)
			if channel == update.ChannelDevelop {
				fmt.Fprintln(cmd.OutOrStdout(), "develop channel — always available via 'harness update --channel develop'")
				return nil
			}
			latest, err := update.PickLatest(channel, rs)
			if err != nil {
				return err
			}
			current := version.Version
			cmp := update.CompareVersions(current, latest.TagName)
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "current\t%s\n", current)
			fmt.Fprintf(w, "channel\t%s\n", channel)
			fmt.Fprintf(w, "latest\t%s\n", latest.TagName)
			fmt.Fprintf(w, "published\t%s\n", latest.PublishedAt.Format("2006-01-02 15:04"))
			switch {
			case cmp < 0:
				fmt.Fprintf(w, "status\tupdate available — run: harness update --channel %s\n", channel)
			case cmp == 0:
				fmt.Fprintln(w, "status\tup to date")
			default:
				fmt.Fprintln(w, "status\tcurrent is newer than channel (development build?)")
			}
			return w.Flush()
		},
	}
	c.Flags().StringVar(&updateFlags.channel, "channel", "stable", "stable|beta|develop")
	c.Flags().StringVar(&updateFlags.repo, "repo", defaultRepo, "GitHub repo (owner/name)")
	return c
}

func newUpdateChannelsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "channels",
		Short: "List releases per channel",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lister := update.NewGitHubLister()
			rs, err := lister.List(updateFlags.repo)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, ch := range update.KnownChannels() {
				if ch == update.ChannelDevelop {
					fmt.Fprintf(out, "%s — source build from git develop branch\n\n", ch)
					continue
				}
				rels := update.FilterChannel(ch, rs)
				if len(rels) == 0 {
					fmt.Fprintf(out, "%s — no releases\n\n", ch)
					continue
				}
				fmt.Fprintf(out, "%s:\n", ch)
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				for _, r := range rels {
					tag := r.TagName
					if r.Prerelease {
						tag += " (prerelease)"
					}
					fmt.Fprintf(w, "  %s\t%s\t%s\n", tag, r.PublishedAt.Format("2006-01-02"), r.HTMLURL)
				}
				_ = w.Flush()
				fmt.Fprintln(out)
			}
			return nil
		},
	}
	c.Flags().StringVar(&updateFlags.repo, "repo", defaultRepo, "GitHub repo (owner/name)")
	return c
}

func runDoUpdate(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	channel := update.Channel(updateFlags.channel)
	if channel == update.ChannelDevelop {
		return updateFromSource(out, updateFlags.repo, updateFlags.dryRun)
	}
	tag := updateFlags.tag
	if tag == "" {
		lister := update.NewGitHubLister()
		rs, err := lister.List(updateFlags.repo)
		if err != nil {
			return err
		}
		latest, err := update.PickLatest(channel, rs)
		if err != nil {
			return err
		}
		tag = latest.TagName
	}
	return updateFromRelease(out, updateFlags.repo, tag, updateFlags.dryRun)
}

func updateFromRelease(out io.Writer, repo, tag string, dryRun bool) error {
	current := normaliseTag(version.Version)
	tag = normaliseTag(tag)
	fmt.Fprintf(out, "current:  %s\ntarget:   %s\nrepo:     %s\n", current, tag, repo)
	cmp := update.CompareVersions(current, tag)
	if cmp == 0 {
		if !updateFlags.force {
			fmt.Fprintln(out, "already on target tag — nothing to do (pass --force to reinstall)")
			return nil
		}
		fmt.Fprintln(out, "→ same tag, --force set: reinstalling")
	}
	if cmp > 0 && !updateFlags.force {
		return fmt.Errorf("refusing downgrade %s → %s (pass --force to override)", current, tag)
	}
	target := update.PlatformTarget()
	url := update.TarballURL(repo, tag, target)
	shaURL := url + ".sha256"
	fmt.Fprintf(out, "→ downloading %s\n", url)

	tmpDir, err := os.MkdirTemp("", "harness-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, target+".tar.gz")
	if err := update.DownloadFile(url, tarPath); err != nil {
		return fmt.Errorf("download tarball: %w", err)
	}
	shaPath := tarPath + ".sha256"
	if err := update.DownloadFile(shaURL, shaPath); err != nil {
		return fmt.Errorf("download sha256: %w", err)
	}

	fmt.Fprintln(out, "→ verifying SHA-256")
	if err := update.VerifySha256(tarPath, shaPath); err != nil {
		return err
	}

	fmt.Fprintln(out, "→ extracting")
	binPath, err := update.ExtractTarget(tarPath, tmpDir, target)
	if err != nil {
		return err
	}

	dest, err := resolveExecutable()
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(out, "→ dry-run: would replace %s with %s\n", dest, binPath)
		return runNew(out, binPath, "version")
	}

	fmt.Fprintf(out, "→ installing %s\n", dest)
	if err := update.ReplaceBinary(binPath, dest); err != nil {
		return err
	}
	fmt.Fprintln(out)
	return runNew(out, dest, "version")
}

func updateFromSource(out io.Writer, repo string, dryRun bool) error {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git required for source channel: %w", err)
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go toolchain required for source channel: %w", err)
	}
	tmpDir, err := os.MkdirTemp("", "harness-src-")
	if err != nil {
		return err
	}
	if !dryRun {
		defer os.RemoveAll(tmpDir)
	}
	url := fmt.Sprintf("https://github.com/%s.git", repo)
	fmt.Fprintf(out, "→ cloning %s → %s\n", url, tmpDir)
	if outBytes, err := runCmd(gitBin, tmpDir, "clone", "--depth", "1", "--branch", "develop", url, "."); err != nil {
		return fmt.Errorf("git clone: %w: %s", err, outBytes)
	}
	fmt.Fprintln(out, "→ building bin/harness")
	if outBytes, err := runCmd(goBin, tmpDir, "build", "-trimpath", "-o", "bin/harness", "./cmd/harness"); err != nil {
		return fmt.Errorf("go build: %w: %s", err, outBytes)
	}
	src := filepath.Join(tmpDir, "bin", "harness")
	dest, err := resolveExecutable()
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(out, "→ dry-run: built %s; would replace %s\n", src, dest)
		return runNew(out, src, "version")
	}
	fmt.Fprintf(out, "→ installing %s\n", dest)
	if err := update.ReplaceBinary(src, dest); err != nil {
		return err
	}
	return runNew(out, dest, "version")
}

func resolveExecutable() (string, error) {
	dest, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current binary: %w", err)
	}
	if real, err := filepath.EvalSymlinks(dest); err == nil {
		dest = real
	}
	return dest, nil
}

func runCmd(bin, dir string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func runNew(out io.Writer, bin, sub string) error {
	cmd := exec.Command(bin, sub)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
