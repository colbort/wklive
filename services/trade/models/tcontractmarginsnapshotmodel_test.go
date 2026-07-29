package models

import (
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMarginRiskProjectionEqual(t *testing.T) {
	a := &TContractMarginSnapshot{
		WalletBalance:     decimal.NewFromInt(100),
		AvailableBalance:  decimal.NewFromInt(70),
		FrozenBalance:     decimal.NewFromInt(30),
		PositionMargin:    decimal.NewFromInt(20),
		OrderMargin:       decimal.NewFromInt(30),
		MaintenanceMargin: decimal.NewFromInt(5),
		AccountEquity:     decimal.NewFromInt(110),
		AvailableMargin:   decimal.NewFromInt(60),
		RiskRate:          decimal.RequireFromString("0.0454545455"),
		PositionCount:     2,
		AssetVersion:      3,
		UnrealizedPnl:     decimal.NewFromInt(-10),
		RealizedPnl:       decimal.NewFromInt(4),
		SourceEventNo:     sql.NullString{String: "CR-a", Valid: true},
		SnapshotTime:      10,
	}
	b := *a
	if !marginRiskProjectionEqual(a, &b) {
		t.Fatal("identical projections should not advance the version")
	}
	b.OrderMargin = decimal.NewFromInt(31)
	if marginRiskProjectionEqual(a, &b) {
		t.Fatal("order margin change must advance the projection")
	}
}
