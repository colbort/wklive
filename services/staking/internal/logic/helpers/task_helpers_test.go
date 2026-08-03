package helpers

import (
	"testing"

	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
)

func TestCalcTaskRewardUses365DayBasisAndEightDecimals(t *testing.T) {
	order := &models.TStakeOrder{StakeAmount: decimal.NewFromInt(1000), Apr: decimal.RequireFromString("12.5")}
	if got, want := CalcTaskReward(order, 1).StringFixed(8), "0.34246575"; got != want {
		t.Fatalf("daily reward = %s, want %s", got, want)
	}
	if got, want := CalcTaskReward(order, 30).StringFixed(8), "10.27397260"; got != want {
		t.Fatalf("30-day reward = %s, want %s", got, want)
	}
}

func TestCalcTaskRewardRejectsInvalidInput(t *testing.T) {
	if !CalcTaskReward(nil, 1).IsZero() {
		t.Fatal("nil order must yield zero")
	}
	if !CalcTaskReward(&models.TStakeOrder{StakeAmount: decimal.NewFromInt(1), Apr: decimal.NewFromInt(1)}, 0).IsZero() {
		t.Fatal("zero days must yield zero")
	}
}
