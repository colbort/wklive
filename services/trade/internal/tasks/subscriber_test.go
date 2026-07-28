package tasks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wklive/common/helper"
	"wklive/common/i18n"
	commontasks "wklive/common/tasks"
	"wklive/proto/trade"
)

func TestUnknownTaskActionIsNotAcknowledged(t *testing.T) {
	err := handleTask(context.Background(), nil, commontasks.Message{Action: "ProcessContractSettlments"})
	if err == nil || !strings.Contains(err.Error(), "unknown trade task action") {
		t.Fatalf("unknown action error=%v", err)
	}
}

func TestAllScheduledTradeActionsHaveSubscriberRoutes(t *testing.T) {
	actions := []string{
		commontasks.ActionTradeProcessOrderMatching,
		commontasks.ActionTradeProcessPositions,
		commontasks.ActionTradeProcessContractSettlements,
		commontasks.ActionTradeProcessSecondsSettlements,
		commontasks.ActionTradeProcessTradeEvents,
		commontasks.ActionTradeExpireRiskLimits,
		commontasks.ActionTradeArchiveLiquidityOrders,
	}
	for _, action := range actions {
		if taskHandlerFor(action) == nil {
			t.Fatalf("scheduled Trade action has no subscriber route: %s", action)
		}
	}
	if taskHandlerFor("ProcessContractSettlments") != nil {
		t.Fatal("misspelled action unexpectedly has a subscriber route")
	}
}

func TestTaskResponseControlsKafkaAcknowledgement(t *testing.T) {
	upstream := errors.New("task failed")
	if err := checkResp(nil, upstream); !errors.Is(err, upstream) {
		t.Fatalf("upstream error=%v", err)
	}
	if err := checkResp(nil, nil); err == nil {
		t.Fatal("empty task response must be retried")
	}
	if err := checkResp(&trade.TradeTaskResp{Base: helper.ErrResp(500, "failed")}, nil); err == nil {
		t.Fatal("rejected task response must be retried")
	}
	if err := checkResp(&trade.TradeTaskResp{Base: helper.ErrResp(i18n.SyncTaskAlreadyRunning, "running")}, nil); err != nil {
		t.Fatalf("already-running duplicate should be acknowledged: %v", err)
	}
	if err := checkResp(&trade.TradeTaskResp{Base: helper.OkResp()}, nil); err != nil {
		t.Fatalf("successful task response rejected: %v", err)
	}
}
