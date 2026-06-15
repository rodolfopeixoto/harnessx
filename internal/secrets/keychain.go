// SPDX-License-Identifier: MIT

package secrets

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const keychainService = "harnessx"

type KeychainBackend struct{}

func (KeychainBackend) Name() string { return "keychain" }

func (KeychainBackend) Available() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath("security")
	return err == nil
}

func (KeychainBackend) Get(name string) (string, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", keychainService, "-a", name, "-w").Output()
	if err != nil {
		return "", ErrNotFound
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func (KeychainBackend) Set(name, value string) error {
	args := []string{"add-generic-password", "-U", "-s", keychainService, "-a", name, "-w", value}
	out, err := exec.Command("security", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain set: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (KeychainBackend) Delete(name string) error {
	out, err := exec.Command("security", "delete-generic-password", "-s", keychainService, "-a", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("keychain delete: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (KeychainBackend) List() ([]string, error) {
	out, err := exec.Command("security", "dump-keychain").Output()
	if err != nil {
		return nil, err
	}
	return parseKeychainDump(string(out)), nil
}

func parseKeychainDump(_ string) []string {
	return nil
}
