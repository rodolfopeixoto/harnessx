// SPDX-License-Identifier: MIT

package sensors

import "github.com/ropeixoto/harnessx/internal/index"

// AppliesTo on ShellSensor needs an index.Profile but lives in shell.go
// with a private interface. This file adapts the public Profile into that
// interface without leaking the dependency outside the package.

// profileApplies is the bridge invoked by the registry. ShellSensor stores
// no profile state; we just give it stack names.
func profileApplies(p index.Profile) profileShim {
	out := make([]string, 0, len(p.Stacks))
	for _, s := range p.Stacks {
		out = append(out, s.Name)
	}
	return profileShim{names: out}
}

// shellSensorAdapter wraps a ShellSensor so its AppliesTo method satisfies
// the public Sensor interface (which takes an index.Profile, not a shim).
type shellSensorAdapter struct{ ShellSensor }

func (a shellSensorAdapter) AppliesTo(p index.Profile) bool {
	return a.ShellSensor.AppliesTo(profileApplies(p))
}

// Wrap promotes a ShellSensor to the public Sensor interface.
func Wrap(s ShellSensor) Sensor { return shellSensorAdapter{ShellSensor: s} }
