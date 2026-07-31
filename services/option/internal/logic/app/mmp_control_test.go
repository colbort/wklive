package applogic

import (
	"strings"
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func activeMMPConfig() *models.TOptionMmpConfig {
	return &models.TOptionMmpConfig{
		Enabled:             int64(common.YesNo_YES_NO_YES),
		Status:              int64(option.MMPStatus_MMP_STATUS_ACTIVE),
		QtyThreshold:        decimal.NewFromInt(10),
		TradeCountThreshold: 3,
		LossThreshold:       decimal.NewFromInt(20),
		WindowSeconds:       5,
		CooldownSeconds:     30,
		WindowStart:         100,
	}
}

func TestNormalizeMMPGroup(t *testing.T) {
	for _, valid := range []string{"maker_1", "desk-A", "ABC123"} {
		if got, ok := NormalizeMMPGroup(" " + valid + " "); !ok || got != strings.ToLower(valid) {
			t.Fatalf("valid group %q normalized to %q ok=%t", valid, got, ok)
		}
	}
	for _, invalid := range []string{"", "a/b", "a b", "中文", "123456789012345678901234567890123"} {
		if _, ok := NormalizeMMPGroup(invalid); ok {
			t.Fatalf("invalid group %q accepted", invalid)
		}
	}
}

func TestApplyMMPFillTriggersAtExactQtyBoundary(t *testing.T) {
	config := activeMMPConfig()
	triggered, reason := applyMMPFill(
		config, int64(common.Side_SIDE_BUY),
		decimal.NewFromInt(100), decimal.NewFromInt(10),
		decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.Zero, 101,
	)
	if !triggered || reason != "QTY_THRESHOLD" {
		t.Fatalf("triggered=%t reason=%q", triggered, reason)
	}
	if config.Status != int64(option.MMPStatus_MMP_STATUS_TRIGGERED) ||
		config.CooldownUntil != 131 {
		t.Fatalf("status=%d cooldown=%d", config.Status, config.CooldownUntil)
	}
}

func TestApplyMMPFillCalculatesBuyAndSellAdverseLoss(t *testing.T) {
	buy := activeMMPConfig()
	buy.QtyThreshold = decimal.Zero
	buy.TradeCountThreshold = 0
	buy.LossThreshold = decimal.RequireFromString("10.5")
	triggered, reason := applyMMPFill(
		buy, int64(common.Side_SIDE_BUY),
		decimal.NewFromInt(105), decimal.NewFromInt(2),
		decimal.NewFromInt(100), decimal.NewFromInt(1),
		decimal.RequireFromString("0.5"), 101,
	)
	if !triggered || reason != "LOSS_THRESHOLD" ||
		!buy.AccumulatedLoss.Equal(decimal.RequireFromString("10.5")) {
		t.Fatalf("buy triggered=%t reason=%q loss=%s", triggered, reason, buy.AccumulatedLoss)
	}

	sell := activeMMPConfig()
	sell.QtyThreshold = decimal.Zero
	sell.TradeCountThreshold = 0
	sell.LossThreshold = decimal.RequireFromString("8.25")
	triggered, reason = applyMMPFill(
		sell, int64(common.Side_SIDE_SELL),
		decimal.NewFromInt(96), decimal.NewFromInt(2),
		decimal.NewFromInt(100), decimal.NewFromInt(1),
		decimal.RequireFromString("0.25"), 101,
	)
	if !triggered || reason != "LOSS_THRESHOLD" ||
		!sell.AccumulatedLoss.Equal(decimal.RequireFromString("8.25")) {
		t.Fatalf("sell triggered=%t reason=%q loss=%s", triggered, reason, sell.AccumulatedLoss)
	}
}

func TestApplyMMPFillResetsExpiredWindow(t *testing.T) {
	config := activeMMPConfig()
	config.AccumulatedQty = decimal.NewFromInt(9)
	config.TradeCount = 2
	config.AccumulatedLoss = decimal.NewFromInt(19)
	triggered, reason := applyMMPFill(
		config, int64(common.Side_SIDE_BUY),
		decimal.NewFromInt(100), decimal.NewFromInt(1),
		decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.Zero, 105,
	)
	if triggered || reason != "" {
		t.Fatalf("expired window must not inherit counters: triggered=%t reason=%q", triggered, reason)
	}
	if config.WindowStart != 105 || !config.AccumulatedQty.Equal(decimal.NewFromInt(1)) ||
		config.TradeCount != 1 || !config.AccumulatedLoss.IsZero() {
		t.Fatalf(
			"window=%d qty=%s count=%d loss=%s",
			config.WindowStart, config.AccumulatedQty, config.TradeCount, config.AccumulatedLoss,
		)
	}
}

func TestApplyMMPFillDoesNotMutateTriggeredConfig(t *testing.T) {
	config := activeMMPConfig()
	config.Status = int64(option.MMPStatus_MMP_STATUS_TRIGGERED)
	triggered, reason := applyMMPFill(
		config, int64(common.Side_SIDE_BUY),
		decimal.NewFromInt(200), decimal.NewFromInt(100),
		decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.NewFromInt(10), 101,
	)
	if triggered || reason != "" || !config.AccumulatedQty.IsZero() {
		t.Fatal("triggered group must reject further accounting")
	}
}
