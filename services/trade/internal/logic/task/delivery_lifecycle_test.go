package tasklogic

import (
	"testing"

	"wklive/proto/trade"
)

func TestDeliveryBatchLifecycleIsMonotonic(t *testing.T) {
	tests := []struct {
		name    string
		current trade.DeliveryBatchStatus
		target  trade.DeliveryBatchStatus
		want    bool
	}{
		{"close only starts lifecycle", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_UNKNOWN, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_CLOSE_ONLY, true},
		{"matching stop follows close only", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_CLOSE_ONLY, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED, true},
		{"price lock follows matching stop", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING, true},
		{"same state is idempotent", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED, false},
		{"scheduler cannot regress", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_CLOSE_ONLY, false},
		{"settling cannot regress", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING, false},
		{"completed cannot regress", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_COMPLETED, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MATCHING_STOPPED, false},
		{"manual review cannot regress", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_MANUAL_REVIEW, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING, false},
		{"lifecycle cannot jump to settling through helper", trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_PRICE_LOCKING, trade.DeliveryBatchStatus_DELIVERY_BATCH_STATUS_SETTLING, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldAdvanceDeliveryBatchStatus(int64(tt.current), tt.target); got != tt.want {
				t.Fatalf("shouldAdvanceDeliveryBatchStatus(%s, %s)=%v want=%v", tt.current, tt.target, got, tt.want)
			}
		})
	}
}
