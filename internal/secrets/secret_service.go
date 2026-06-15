// SPDX-License-Identifier: MIT

package secrets

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const secretServiceLabel = "harnessx"

type SecretServiceBackend struct{}

func (SecretServiceBackend) Name() string { return "secret_service" }

func (SecretServiceBackend) Available() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("secret-tool")
	return err == nil
}

func (SecretServiceBackend) Get(name string) (string, error) {
	out, err := exec.Command("secret-tool", "lookup", "service", secretServiceLabel, "account", name).Output()
	if err != nil {
		return "", ErrNotFound
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (SecretServiceBackend) Set(name, value string) error {
	cmd := exec.Command("secret-tool", "store", "--label", secretServiceLabel+":"+name, "service", secretServiceLabel, "account", name)
	cmd.Stdin = strings.NewReader(value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("secret_service set: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (SecretServiceBackend) Delete(name string) error {
	out, err := exec.Command("secret-tool", "clear", "service", secretServiceLabel, "account", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("secret_service delete: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (SecretServiceBackend) List() ([]string, error) {
	return nil, nil
}
