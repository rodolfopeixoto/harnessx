// SPDX-License-Identifier: MIT

//go:build !windows

package repl

import (
	"os"
	"os/signal"
	"syscall"
)

func (r *readlinePromptReader) startWinch() {
	r.winch = make(chan os.Signal, 1)
	r.closed = make(chan struct{})
	signal.Notify(r.winch, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-r.closed:
				return
			case <-r.winch:
				if r.rl != nil {
					r.rl.Refresh()
				}
			}
		}
	}()
}

func (r *readlinePromptReader) stopWinch() {
	if r.winch == nil {
		return
	}
	signal.Stop(r.winch)
	close(r.closed)
}
