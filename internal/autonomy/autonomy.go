// SPDX-License-Identifier: MIT

package autonomy

import (
	"errors"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

type Level string

const (
	Manual               = Level(constants.AutonomyManual)
	PlanAndAsk           = Level(constants.AutonomyPlanAndAsk)
	SafeExecute          = Level(constants.AutonomySafeExecute)
	FullProjectLoop      = Level(constants.AutonomyFullProjectLoop)
	ScheduledMaintenance = Level(constants.AutonomyScheduledMaintenance)
)

func AllLevels() []Level {
	return []Level{Manual, PlanAndAsk, SafeExecute, FullProjectLoop, ScheduledMaintenance}
}

type Operation string

const (
	OpRead            = Operation(constants.AutonomyOpRead)
	OpPlan            = Operation(constants.AutonomyOpPlan)
	OpExecuteLowRisk  = Operation(constants.AutonomyOpExecuteLowRisk)
	OpExecuteHighRisk = Operation(constants.AutonomyOpExecuteHighRisk)
	OpClean           = Operation(constants.AutonomyOpClean)
	OpSchedule        = Operation(constants.AutonomyOpSchedule)
)

type Decision string

const (
	DecisionAllow    Decision = "allow"
	DecisionApproval Decision = "require_approval"
	DecisionDeny     Decision = "deny"
)

var ErrUnknownLevel = errors.New("autonomy: unknown level")

var policy = map[Level]map[Operation]Decision{
	Manual: {
		OpRead:            DecisionAllow,
		OpPlan:            DecisionApproval,
		OpExecuteLowRisk:  DecisionDeny,
		OpExecuteHighRisk: DecisionDeny,
		OpClean:           DecisionDeny,
		OpSchedule:        DecisionDeny,
	},
	PlanAndAsk: {
		OpRead:            DecisionAllow,
		OpPlan:            DecisionAllow,
		OpExecuteLowRisk:  DecisionApproval,
		OpExecuteHighRisk: DecisionApproval,
		OpClean:           DecisionApproval,
		OpSchedule:        DecisionApproval,
	},
	SafeExecute: {
		OpRead:            DecisionAllow,
		OpPlan:            DecisionAllow,
		OpExecuteLowRisk:  DecisionAllow,
		OpExecuteHighRisk: DecisionApproval,
		OpClean:           DecisionApproval,
		OpSchedule:        DecisionAllow,
	},
	FullProjectLoop: {
		OpRead:            DecisionAllow,
		OpPlan:            DecisionAllow,
		OpExecuteLowRisk:  DecisionAllow,
		OpExecuteHighRisk: DecisionApproval,
		OpClean:           DecisionApproval,
		OpSchedule:        DecisionAllow,
	},
	ScheduledMaintenance: {
		OpRead:            DecisionAllow,
		OpPlan:            DecisionAllow,
		OpExecuteLowRisk:  DecisionApproval,
		OpExecuteHighRisk: DecisionDeny,
		OpClean:           DecisionApproval,
		OpSchedule:        DecisionAllow,
	},
}

func Gate(level Level, op Operation) (Decision, error) {
	ops, ok := policy[level]
	if !ok {
		return DecisionDeny, ErrUnknownLevel
	}
	if d, ok := ops[op]; ok {
		return d, nil
	}
	return DecisionDeny, nil
}

type Setting struct {
	Level Level `json:"level"`
}

func DefaultSetting() Setting {
	return Setting{Level: PlanAndAsk}
}
