// SPDX-License-Identifier: MIT

package clock

import "time"

type Clock interface {
	Now() time.Time
}

type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

type Fake struct {
	T time.Time
}

func (f *Fake) Now() time.Time { return f.T }

func (f *Fake) Advance(d time.Duration) { f.T = f.T.Add(d) }
