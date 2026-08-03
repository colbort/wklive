package helpers

const (
	StakeOperationTypeSubscribe      int64 = 1
	StakeOperationTypeDailyReward    int64 = 2
	StakeOperationTypeMaturityReward int64 = 3
	StakeOperationTypeMaturityRedeem int64 = 4
	StakeOperationTypeEarlyRedeem    int64 = 5
	StakeOperationTypeManualReward   int64 = 6
	StakeOperationTypeManualRedeem   int64 = 7
	StakeOperationTypeSubscribeUndo  int64 = 8
)

const (
	StakeOperationStepNotRequired int64 = 0
	StakeOperationStepPending     int64 = 1
	StakeOperationStepSucceeded   int64 = 2
)

const (
	StakeOperationStatusPending         int64 = 1
	StakeOperationStatusProcessing      int64 = 2
	StakeOperationStatusSucceeded       int64 = 3
	StakeOperationStatusRetryableFailed int64 = 4
	StakeOperationStatusManualRequired  int64 = 5
)

const stakeOperationMaxRetries int64 = 20

func StakeOperationRetryStatus(retryCount int64) int64 {
	if retryCount >= stakeOperationMaxRetries {
		return StakeOperationStatusManualRequired
	}
	return StakeOperationStatusRetryableFailed
}

func StakeOperationStepBizNo(operationNo, step string) string {
	return operationNo + ":" + step
}
