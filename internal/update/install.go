// SPDX-License-Identifier: MIT

package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const tarballSizeLimit = 200 * 1024 * 1024

func PlatformTarget() string {
	return fmt.Sprintf("harness-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func TarballURL(repo, tag, target string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s.tar.gz", repo, tag, target)
}

func DownloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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

func VerifySha256(tarPath, shaPath string) error {
	expectedBytes, err := os.ReadFile(shaPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(strings.TrimSpace(string(expectedBytes)))
	if len(fields) == 0 {
		return fmt.Errorf("sha256: empty checksum file")
	}
	expected := fields[0]
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

func ExtractTarget(tarPath, tmpDir, target string) (string, error) {
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
		if _, err := io.CopyN(w, tr, tarballSizeLimit); err != nil && err != io.EOF {
			w.Close()
			return "", err
		}
		w.Close()
		return out, nil
	}
	return "", fmt.Errorf("binary %s not found in tarball", target)
}

func ReplaceBinary(src, dest string) error {
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
