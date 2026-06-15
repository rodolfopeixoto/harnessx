// SPDX-License-Identifier: MIT
//go:build !windows

package interactive

import (
	"errors"
	"io"
	"os"
	"syscall"
	"time"
)

func setReadDeadline(r io.Reader, t time.Time) error {
	if f, ok := r.(*os.File); ok {
		return f.SetReadDeadline(t)
	}
	return nil
}

func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var sysErr syscall.Errno
	return errors.As(err, &sysErr) && sysErr == syscall.EAGAIN
}
