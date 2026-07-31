package applogic

import (
	"errors"
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

func TestComboRequestPayloadHashIsCanonicalAndSensitive(t *testing.T) {
	base := &option.PlaceComboOrderReq{
		AccountId: 10, ClientComboId: "combo-001",
		OrderType: option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT,
		Qty:       "2", NetPrice: "3",
		Legs: []*option.ComboOrderLegInput{
			{ContractId: 200, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "7"},
			{ContractId: 100, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "10"},
		},
	}
	hash, err := comboRequestPayloadHash(base)
	if err != nil {
		t.Fatalf("hash base request: %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("hash length=%d, want 64", len(hash))
	}
	reordered := proto.Clone(base).(*option.PlaceComboOrderReq)
	reordered.Legs = []*option.ComboOrderLegInput{base.Legs[1], base.Legs[0]}
	reorderedHash, err := comboRequestPayloadHash(reordered)
	if err != nil {
		t.Fatalf("hash reordered request: %v", err)
	}
	if reorderedHash != hash {
		t.Fatalf("canonical hash differs after leg reordering: %s != %s", reorderedHash, hash)
	}
	changed := proto.Clone(base).(*option.PlaceComboOrderReq)
	changed.NetPrice = "4"
	changed.Legs = []*option.ComboOrderLegInput{
		base.Legs[0],
		{ContractId: 100, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "11"},
	}
	changedHash, err := comboRequestPayloadHash(changed)
	if err != nil {
		t.Fatalf("hash changed request: %v", err)
	}
	if changedHash == hash {
		t.Fatal("economic payload change did not change hash")
	}
}

func TestComboRequestPayloadHashRejectsInvalidEconomics(t *testing.T) {
	tests := []struct {
		name string
		req  *option.PlaceComboOrderReq
	}{
		{
			name: "net price mismatch",
			req: &option.PlaceComboOrderReq{
				AccountId: 1, ClientComboId: "bad-net",
				OrderType: option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT,
				Qty:       "1", NetPrice: "99",
				Legs: []*option.ComboOrderLegInput{
					{ContractId: 1, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "10"},
					{ContractId: 2, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "7"},
				},
			},
		},
		{
			name: "ratio not reduced",
			req: &option.PlaceComboOrderReq{
				AccountId: 1, ClientComboId: "bad-ratio",
				OrderType: option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT,
				Qty:       "1", NetPrice: "6",
				Legs: []*option.ComboOrderLegInput{
					{ContractId: 1, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 2, Price: "10"},
					{ContractId: 2, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 2, Price: "7"},
				},
			},
		},
		{
			name: "duplicate contract",
			req: &option.PlaceComboOrderReq{
				AccountId: 1, ClientComboId: "duplicate",
				OrderType: option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT,
				Qty:       "1", NetPrice: "3",
				Legs: []*option.ComboOrderLegInput{
					{ContractId: 1, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "10"},
					{ContractId: 1, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "7"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := comboRequestPayloadHash(test.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateInverseComboLegsAndProgress(t *testing.T) {
	incomingParent := &models.TOptionComboOrder{Id: 11}
	makerParent := &models.TOptionComboOrder{Id: 22}
	incomingLeg := &models.TOptionComboOrderLeg{
		Id: 1, ComboOrderId: 11, LegNo: 1, ContractId: 100,
		Side: int64(common.Side_SIDE_BUY), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		Ratio: 2, Price: decimal.RequireFromString("10"),
		Qty: decimal.RequireFromString("6"), UnfilledQty: decimal.RequireFromString("6"),
		ChildOrderId: 101,
	}
	makerLeg := &models.TOptionComboOrderLeg{
		Id: 2, ComboOrderId: 22, LegNo: 1, ContractId: 100,
		Side: int64(common.Side_SIDE_SELL), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		Ratio: 2, Price: decimal.RequireFromString("9"),
		Qty: decimal.RequireFromString("6"), UnfilledQty: decimal.RequireFromString("6"),
		ChildOrderId: 202,
	}
	incomingChild := &models.TOptionOrder{
		Id: 101, ComboOrderId: 11, ComboLegNo: 1, ContractId: 100,
		Side: int64(common.Side_SIDE_BUY), Price: decimal.RequireFromString("10"),
		Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
	}
	makerChild := &models.TOptionOrder{
		Id: 202, ComboOrderId: 22, ComboLegNo: 1, ContractId: 100,
		Side: int64(common.Side_SIDE_SELL), Price: decimal.RequireFromString("9"),
		Status: int64(option.OrderStatus_ORDER_STATUS_PENDING),
	}
	if err := validateInverseComboLegs(
		incomingParent, makerParent, incomingLeg, makerLeg, incomingChild, makerChild,
	); err != nil {
		t.Fatalf("valid inverse legs rejected: %v", err)
	}
	applyComboLegFill(incomingLeg, decimal.RequireFromString("2"), 123)
	if !incomingLeg.FilledQty.Equal(decimal.RequireFromString("2")) ||
		!incomingLeg.UnfilledQty.Equal(decimal.RequireFromString("4")) {
		t.Fatalf("unexpected leg progress: filled=%s unfilled=%s", incomingLeg.FilledQty, incomingLeg.UnfilledQty)
	}
	parent := &models.TOptionComboOrder{
		Qty: decimal.RequireFromString("3"), UnfilledQty: decimal.RequireFromString("3"),
		Status: int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE),
	}
	applyComboParentFill(parent, decimal.RequireFromString("1"), 123)
	if parent.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED) ||
		!parent.UnfilledQty.Equal(decimal.RequireFromString("2")) {
		t.Fatalf("unexpected partial parent progress: status=%d unfilled=%s", parent.Status, parent.UnfilledQty)
	}
	applyComboParentFill(parent, decimal.RequireFromString("2"), 124)
	if parent.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_FILLED) ||
		!parent.UnfilledQty.IsZero() {
		t.Fatalf("unexpected filled parent progress: status=%d unfilled=%s", parent.Status, parent.UnfilledQty)
	}
	makerLeg.Side = int64(common.Side_SIDE_BUY)
	if err := validateInverseComboLegs(
		incomingParent, makerParent, incomingLeg, makerLeg, incomingChild, makerChild,
	); err == nil {
		t.Fatal("same-side legs were accepted")
	}
	makerLeg.Side = int64(common.Side_SIDE_SELL)
	makerLeg.Price = decimal.RequireFromString("11")
	if err := validateInverseComboLegs(
		incomingParent, makerParent, incomingLeg, makerLeg, incomingChild, makerChild,
	); !errors.Is(err, errComboLegLimitNotExecutable) {
		t.Fatalf("buy-leg price cap result=%v, want limit sentinel", err)
	}
}
