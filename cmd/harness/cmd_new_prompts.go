// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// promptChoice reads a line from in and validates it against a fixed set of
// options — used by the interactive `harness new` wizard.
func promptChoice(in io.Reader, out io.Writer, label string, options []string) (string, error) {
	fmt.Fprintf(out, "%s? (%s)\n> ", label, strings.Join(options, "|"))
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if !contains(options, line) {
		return "", fmt.Errorf("invalid choice %q", line)
	}
	return line, nil
}

func promptString(in io.Reader, out io.Writer, label, fallback string) (string, error) {
	fmt.Fprintf(out, "%s? [%s]\n> ", label, fallback)
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return fallback, nil
	}
	return line, nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
