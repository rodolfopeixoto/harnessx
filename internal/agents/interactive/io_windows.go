// SPDX-License-Identifier: MIT
//go:build windows

package interactive

import (
	"io"
	"time"
)

func setReadDeadline(_ io.Reader, _ time.Time) error { return nil }

func isTimeoutErr(_ error) bool { return false }
