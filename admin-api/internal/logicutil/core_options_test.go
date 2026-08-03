package logicutil

import "testing"

func TestStakingOptionsContainAllBusinessEnums(t *testing.T) {
	want := map[string]bool{
		"productStatus": false, "stakingProductType": false, "interestMode": false,
		"rewardMode": false, "stakingOrderStatus": false, "stakingRedeemType": false,
		"stakingRewardType": false, "stakingRewardStatus": false, "stakingRedeemStatus": false,
		"stakingSourceType": false, "stakingOperationType": false,
		"stakingOperationStatus": false, "stakingOperationStepStatus": false,
		"stakingReconciliationStatus": false,
	}
	for _, group := range StakingOptions() {
		if _, ok := want[group.Key]; ok {
			want[group.Key] = len(group.Options) > 0
		}
	}
	for key, populated := range want {
		if !populated {
			t.Fatalf("staking option group %q is missing or empty", key)
		}
	}
}
