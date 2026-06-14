// SPDX-License-Identifier: MIT

package domain

import "time"

type SensorStatus string

const (
	SensorPassed  SensorStatus = "passed"
	SensorFailed  SensorStatus = "failed"
	SensorSkipped SensorStatus = "skipped"
)

type SensorResult struct {
	ID         int64
	RunID      string
	Sensor     string
	Status     SensorStatus
	DurationMs int64
	OutputPath string
	CreatedAt  time.Time
}
