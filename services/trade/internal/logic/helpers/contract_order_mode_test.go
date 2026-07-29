package helpers

import (
	"strings"
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func TestValidateContractOpeningPreference(t *testing.T) {
	preference := &models.TContractUserConfig{
		MarginMode:   int64(trade.MarginMode_MARGIN_MODE_CROSS),
		PositionMode: int64(trade.PositionMode_POSITION_MODE_ONE_WAY),
	}
	if err := validateContractOpeningPreference(
		preference,
		trade.MarginMode_MARGIN_MODE_CROSS,
		trade.PositionMode_POSITION_MODE_ONE_WAY,
	); err != nil {
		t.Fatalf("matching preference rejected: %v", err)
	}
	if err := validateContractOpeningPreference(
		preference,
		trade.MarginMode_MARGIN_MODE_ISOLATED,
		trade.PositionMode_POSITION_MODE_ONE_WAY,
	); err == nil || !strings.Contains(err.Error(), "margin mode") {
		t.Fatalf("margin mismatch error=%v", err)
	}
	if err := validateContractOpeningPreference(
		preference,
		trade.MarginMode_MARGIN_MODE_CROSS,
		trade.PositionMode_POSITION_MODE_HEDGE,
	); err == nil || !strings.Contains(err.Error(), "position mode") {
		t.Fatalf("position mismatch error=%v", err)
	}
	if err := validateContractOpeningPreference(
		nil,
		trade.MarginMode_MARGIN_MODE_ISOLATED,
		trade.PositionMode_POSITION_MODE_HEDGE,
	); err != nil {
		t.Fatalf("missing optional preference rejected: %v", err)
	}
}

func TestContractPositionMode(t *testing.T) {
	tests := []struct {
		side trade.PositionSide
		want trade.PositionMode
	}{
		{trade.PositionSide_POSITION_SIDE_NET, trade.PositionMode_POSITION_MODE_ONE_WAY},
		{trade.PositionSide_POSITION_SIDE_LONG, trade.PositionMode_POSITION_MODE_HEDGE},
		{trade.PositionSide_POSITION_SIDE_SHORT, trade.PositionMode_POSITION_MODE_HEDGE},
	}
	for _, tt := range tests {
		got, err := ContractPositionMode(tt.side)
		if err != nil || got != tt.want {
			t.Fatalf("side=%v got=%v err=%v want=%v", tt.side, got, err, tt.want)
		}
	}
	if _, err := ContractPositionMode(trade.PositionSide_POSITION_SIDE_UNKNOWN); err == nil {
		t.Fatal("unknown position side was accepted")
	}
}
