package tasklogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func TestAssetTimeoutTransitionsToDurableRetry(t *testing.T) {
	const now = int64(1_000_000)
	status, next := settlementFailureTransition(1, now)
	if status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED) {
		t.Fatalf("status=%d, want FAILED", status)
	}
	if next != now+2_000 {
		t.Fatalf("next_retry_at=%d, want %d", next, now+2_000)
	}
}

func TestRepeatedAssetTimeoutBecomesLocatableManualReview(t *testing.T) {
	status, next := settlementFailureTransition(spotSettlementMaxRetry, 1_000_000)
	if status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW) || next != 0 {
		t.Fatalf("status=%d next=%d, want MANUAL_REVIEW with no automatic retry", status, next)
	}
}

func TestSettlementRetryBackoffIsCapped(t *testing.T) {
	const now = int64(1_000_000)
	_, next10 := settlementFailureTransition(10, now)
	_, next19 := settlementFailureTransition(19, now)
	if next10 != next19 || next10 != now+1_024_000 {
		t.Fatalf("backoff cap differs: retry10=%d retry19=%d", next10, next19)
	}
}

func TestExpiredWorkerCannotConfirmAssetSuccessAfterLeaseReclaim(t *testing.T) {
	stale := &models.TTradeSettlementInstruction{
		Status:      int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING),
		UpdateTimes: 100,
	}
	reclaimed := &models.TTradeSettlementInstruction{
		Status:      int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING),
		UpdateTimes: 200,
	}
	if settlementInstructionLeaseOwned(reclaimed, stale) {
		t.Fatal("stale worker retained ownership after another instance reclaimed the instruction")
	}
}

func TestUnboundLegacyMarginCannotReachAssetExecutor(t *testing.T) {
	item := &models.TTradeSettlementInstruction{
		Action:     int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN),
		FillId:     10,
		PositionId: 0,
	}
	if !isUnboundProjectedMarginInstruction(item) {
		t.Fatal("legacy unbound margin instruction was not intercepted")
	}
	item.PositionId = 20
	if isUnboundProjectedMarginInstruction(item) {
		t.Fatal("properly bound margin instruction was incorrectly intercepted")
	}
}
