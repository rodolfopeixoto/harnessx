// SPDX-License-Identifier: MIT

package cleanup

import "github.com/ropeixoto/harnessx/internal/platform/constants"

const (
	RiskLow    Risk = Risk(constants.CleanupRiskLow)
	RiskMedium Risk = Risk(constants.CleanupRiskMedium)
	RiskHigh   Risk = Risk(constants.CleanupRiskHigh)
)
