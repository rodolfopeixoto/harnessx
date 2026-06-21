// SPDX-License-Identifier: MIT

//go:build windows

package repl

func (r *readlinePromptReader) startWinch() {}

func (r *readlinePromptReader) stopWinch() {}
