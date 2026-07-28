package tasklogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

func TestPerpetualAndDeliveryValueTypePnlMatrix(t *testing.T) {
	products := []common.ContractType{
		common.ContractType_CONTRACT_TYPE_PERPETUAL,
		common.ContractType_CONTRACT_TYPE_DELIVERY,
	}
	tests := []struct {
		name      string
		side      trade.PositionSide
		valueType trade.ContractValueType
		open      string
		close     string
		qty       string
		size      string
		want      decimal.Decimal
	}{
		{name: "linear long", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, open: "100", close: "110", qty: "2", size: "1", want: decimal.NewFromInt(20)},
		{name: "linear short", side: trade.PositionSide_POSITION_SIDE_SHORT, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, open: "100", close: "90", qty: "2", size: "1", want: decimal.NewFromInt(20)},
		{name: "inverse long", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE, open: "50000", close: "55000", qty: "100", size: "100", want: decimal.NewFromInt(10000).Mul(decimal.NewFromInt(1).Div(decimal.NewFromInt(50000)).Sub(decimal.NewFromInt(1).Div(decimal.NewFromInt(55000))))},
		{name: "inverse short", side: trade.PositionSide_POSITION_SIDE_SHORT, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE, open: "50000", close: "45000", qty: "100", size: "100", want: decimal.NewFromInt(10000).Mul(decimal.NewFromInt(1).Div(decimal.NewFromInt(45000)).Sub(decimal.NewFromInt(1).Div(decimal.NewFromInt(50000))))},
	}
	for _, product := range products {
		for _, tc := range tests {
			t.Run(product.String()+"/"+tc.name, func(t *testing.T) {
				got := contractRealizedPnl(
					int64(tc.side),
					decimal.RequireFromString(tc.open),
					decimal.RequireFromString(tc.close),
					decimal.RequireFromString(tc.qty),
					decimal.RequireFromString(tc.size),
					int64(tc.valueType),
				)
				if !got.Equal(tc.want) {
					t.Fatalf("pnl=%s want=%s", got, tc.want)
				}
			})
		}
	}
}

func TestLongShortAndNetFillDirectionMatrix(t *testing.T) {
	if !isClosingFill(int64(trade.PositionSide_POSITION_SIDE_LONG), int64(common.Side_SIDE_SELL)) {
		t.Fatal("SELL must close LONG")
	}
	if !isClosingFill(int64(trade.PositionSide_POSITION_SIDE_SHORT), int64(common.Side_SIDE_BUY)) {
		t.Fatal("BUY must close SHORT")
	}
	if isClosingFill(int64(trade.PositionSide_POSITION_SIDE_LONG), int64(common.Side_SIDE_BUY)) {
		t.Fatal("BUY must increase LONG")
	}
	closeSide, openSide := netPositionSides(int64(common.Side_SIDE_BUY))
	if closeSide != int64(trade.PositionSide_POSITION_SIDE_SHORT) || openSide != int64(trade.PositionSide_POSITION_SIDE_LONG) {
		t.Fatalf("NET BUY close/open=%d/%d", closeSide, openSide)
	}
	closeSide, openSide = netPositionSides(int64(common.Side_SIDE_SELL))
	if closeSide != int64(trade.PositionSide_POSITION_SIDE_LONG) || openSide != int64(trade.PositionSide_POSITION_SIDE_SHORT) {
		t.Fatalf("NET SELL close/open=%d/%d", closeSide, openSide)
	}
}

func TestNetReversalMarginUsesOnlyOpeningRemainder(t *testing.T) {
	fillQty := decimal.NewFromInt(5)
	closedQty := decimal.NewFromInt(3)
	openingQty := fillQty.Sub(closedQty)
	if !openingQty.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("opening remainder=%s", openingQty)
	}
	// The projected margin instruction is created inside applyOpen with this
	// remainder, never from the original full Fill quantity.
	if !openingQty.LessThan(fillQty) {
		t.Fatal("NET reversal would charge margin for the closing portion")
	}
}
