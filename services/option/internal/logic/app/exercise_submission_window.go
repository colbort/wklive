package applogic

import (
	"math"

	"wklive/services/option/models"
)

const millisecondsPerSecond int64 = 1000

// exerciseSubmissionWindowOpen compares a millisecond wall-clock instant with
// the contract's published second-aligned listing and cutoff timestamps. The
// cutoff is exclusive: cutoff-1ms is valid, while cutoff and every later
// instant are rejected. Invalid or overflowing contract timestamps fail closed.
func exerciseSubmissionWindowOpen(contract *models.TOptionContract, nowMillis int64) bool {
	if contract == nil || contract.ExerciseCutoffTime <= 0 ||
		contract.ListTime > math.MaxInt64/millisecondsPerSecond ||
		contract.ExerciseCutoffTime > math.MaxInt64/millisecondsPerSecond {
		return false
	}
	listMillis := contract.ListTime * millisecondsPerSecond
	cutoffMillis := contract.ExerciseCutoffTime * millisecondsPerSecond
	return nowMillis >= listMillis && nowMillis < cutoffMillis
}
