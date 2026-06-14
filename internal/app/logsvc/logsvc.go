// SPDX-License-Identifier: MIT

// Package logsvc tails the JSONL event log written by other app services.
package logsvc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

type Options struct {
	Path string
	Tail int // most recent N entries; <=0 means all
}

func Print(opts Options, out io.Writer) error {
	if opts.Path == "" {
		return errors.New("logsvc: empty path")
	}
	f, err := os.Open(opts.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(out, "no log file at %s yet\n", opts.Path)
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 4*1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	start := 0
	if opts.Tail > 0 && len(lines) > opts.Tail {
		start = len(lines) - opts.Tail
	}
	for _, l := range lines[start:] {
		fmt.Fprintln(out, l)
	}
	return nil
}
