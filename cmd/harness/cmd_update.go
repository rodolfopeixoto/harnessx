// SPDX-License-Identifier: MIT

package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ropeixoto/harnessx/internal/version"
)

const defaultRepo = "rodolfopeixoto/harnessx"

func newUpdateCmd() *cobra.Command {
	var (
		repo    string
		tag     string
		dryRun  bool
		channel string
	)
	c := &cobra.Command{
		Use:   "update",
		Short: "Self-update to the latest GitHub release",
		Long: `Download the latest harness binary from GitHub releases, verify
SHA-256, and replace the running binary in place. Use --channel develop
to build from source instead (requires git + go).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if channel == "develop" {
				return updateFromSource(out, repo, dryRun)
			}
			return updateFromRelease(out, repo, tag, dryRun)
		},
	}
	c.Flags().StringVar(&repo, "repo", defaultRepo, "GitHub repo (owner/name)")
	c.Flags().StringVar(&tag, "tag", "", "specific tag (default: latest)")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print plan without replacing binary")
	c.Flags().StringVar(&channel, "channel", "release", "release|develop")
	return c
}

func updateFromRelease(out io.Writer, repo, tag string, dryRun bool) error {
	if tag == "" {
		latest, err := resolveLatestTag(repo)
		if err != nil {
			return err
		}
		tag = latest
	}
	current := version.Version
	fmt.Fprintf(out, "current:  %s\ntarget:   %s\nrepo:     %s\n", current, tag, repo)
	if normalizeVer(current) == normalizeVer(tag) {
		fmt.Fprintln(out, "already on target tag — nothing to do")
		return nil
	}
	target := fmt.Sprintf("harness-%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.tar.gz", repo, tag, target)
	shaURL := url + ".sha256"
	fmt.Fprintf(out, "→ downloading %s\n", url)

	tmpDir, err := os.MkdirTemp("", "harness-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, target+".tar.gz")
	if err := downloadFile(url, tarPath); err != nil {
		return fmt.Errorf("download tarball: %w", err)
	}
	shaPath := tarPath + ".sha256"
	if err := downloadFile(shaURL, shaPath); err != nil {
		return fmt.Errorf("download sha256: %w", err)
	}

	fmt.Fprintln(out, "→ verifying SHA-256")
	if err := verifySha256(tarPath, shaPath); err != nil {
		return err
	}

	fmt.Fprintln(out, "→ extracting")
	binPath, err := extractHarness(tarPath, tmpDir, target)
	if err != nil {
		return err
	}

	dest, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current binary: %w", err)
	}
	if real, err := filepath.EvalSymlinks(dest); err == nil {
		dest = real
	}

	if dryRun {
		fmt.Fprintf(out, "→ dry-run: would replace %s with %s\n", dest, binPath)
		return runNew(out, binPath, "version")
	}

	fmt.Fprintf(out, "→ installing %s\n", dest)
	if err := replaceBinary(binPath, dest); err != nil {
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
	dest, err := os.Executable()
	if err != nil {
		return err
	}
	if real, err := filepath.EvalSymlinks(dest); err == nil {
		dest = real
	}
	if dryRun {
		fmt.Fprintf(out, "→ dry-run: built %s; would replace %s\n", src, dest)
		return runNew(out, src, "version")
	}
	fmt.Fprintf(out, "→ installing %s\n", dest)
	if err := replaceBinary(src, dest); err != nil {
		return err
	}
	return runNew(out, dest, "version")
}

func resolveLatestTag(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api %s: status %d", url, resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("no tag_name in github response")
	}
	return payload.TagName, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func verifySha256(tarPath, shaPath string) error {
	expectedBytes, err := os.ReadFile(shaPath)
	if err != nil {
		return err
	}
	expected := strings.Fields(strings.TrimSpace(string(expectedBytes)))[0]
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(expected, got) {
		return fmt.Errorf("sha256 mismatch: expected %s got %s", expected, got)
	}
	return nil
}

func extractHarness(tarPath, tmpDir, target string) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != target {
			continue
		}
		out := filepath.Join(tmpDir, target)
		w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.CopyN(w, tr, 200*1024*1024); err != nil && err != io.EOF {
			w.Close()
			return "", err
		}
		w.Close()
		return out, nil
	}
	return "", fmt.Errorf("binary %s not found in tarball", target)
}

func replaceBinary(src, dest string) error {
	info, err := os.Stat(dest)
	if err != nil {
		return err
	}
	if err := copyFile(src, dest+".new"); err != nil {
		return err
	}
	if err := os.Chmod(dest+".new", info.Mode()); err != nil {
		_ = os.Remove(dest + ".new")
		return err
	}
	if err := os.Rename(dest+".new", dest); err != nil {
		return fmt.Errorf("install: rename: %w (try sudo harness update)", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
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

func normalizeVer(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if i := strings.Index(s, " "); i > 0 {
		s = s[:i]
	}
	if i := strings.Index(s, "-dev"); i > 0 {
		s = s[:i]
	}
	return s
}
