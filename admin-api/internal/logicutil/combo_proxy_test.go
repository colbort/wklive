package logicutil

import (
	"reflect"
	"testing"

	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/option"
)

func TestCopyValuePreservesAdminComboDrillDown(t *testing.T) {
	source := &option.GetAdminComboOrderResp{
		Base: &common.RespBase{Code: 0, Msg: "ok"},
		Data: &option.OptionAdminComboOrderDetail{
			ComboOrder: &option.OptionComboOrder{
				Id: 91, TenantId: 9, ComboNo: "COMBO-91", Status: option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE,
			},
			Legs: []*option.OptionComboOrderLeg{{
				LegNo: 1, ContractId: 101, ChildOrderId: 201,
			}},
			ChildOrders: []*option.OptionOrderDetail{{
				Order: &option.OptionOrder{Id: 201, ComboOrderId: 91, ComboLegNo: 1},
			}},
			Trades: []*option.OptionTradeDetail{{
				Trade: &option.OptionTrade{
					Id: 301, ComboMatchNo: "COMBO-MATCH-1", ComboLegNo: 1,
				},
			}},
			AssetInstructions: []*option.OptionAssetInstruction{{
				Id: 401, OrderId: 201, InstructionNo: "ORDER-201-FREEZE",
			}},
			TradeTotal: 1, AssetInstructionTotal: 1, DataTruncated: true,
		},
	}
	var target types.GetAdminComboOrderResp
	if err := copyValue(reflect.ValueOf(&target), reflect.ValueOf(source)); err != nil {
		t.Fatal(err)
	}
	if target.Code != 0 || target.Msg != "ok" {
		t.Fatalf("base=%d/%q", target.Code, target.Msg)
	}
	if target.Data.ComboOrder.Id != 91 || target.Data.ComboOrder.ComboNo != "COMBO-91" {
		t.Fatalf("combo=%+v", target.Data.ComboOrder)
	}
	if len(target.Data.Legs) != 1 || target.Data.Legs[0].ChildOrderId != 201 {
		t.Fatalf("legs=%+v", target.Data.Legs)
	}
	if len(target.Data.ChildOrders) != 1 ||
		target.Data.ChildOrders[0].Order.ComboOrderId != 91 {
		t.Fatalf("children=%+v", target.Data.ChildOrders)
	}
	if len(target.Data.Trades) != 1 ||
		target.Data.Trades[0].Trade.ComboMatchNo != "COMBO-MATCH-1" {
		t.Fatalf("trades=%+v", target.Data.Trades)
	}
	if len(target.Data.AssetInstructions) != 1 ||
		target.Data.AssetInstructions[0].InstructionNo != "ORDER-201-FREEZE" {
		t.Fatalf("instructions=%+v", target.Data.AssetInstructions)
	}
	if target.Data.TradeTotal != 1 || target.Data.AssetInstructionTotal != 1 ||
		!target.Data.DataTruncated {
		t.Fatalf("totals=%+v", target.Data)
	}
}

func TestCopyValuePreservesComboOperationsWatermarks(t *testing.T) {
	source := &option.GetOperationsOverviewResp{
		Base: &common.RespBase{Code: 0, Msg: "ok"},
		Data: &option.OptionOperationsOverview{
			GeneratedAt:                    1000,
			RiskStaleSeconds:               60,
			ComboStaleSeconds:              45,
			ComboStaleCount:                2,
			ComboManualReviewCount:         1,
			OldestComboExceptionTime:       901,
			ComboInvariantIssueCount:       3,
			ComboIncompleteMatchGroupCount: 4,
		},
	}
	var target types.GetOperationsOverviewResp
	if err := copyValue(reflect.ValueOf(&target), reflect.ValueOf(source)); err != nil {
		t.Fatal(err)
	}
	if target.Data.ComboStaleSeconds != 45 ||
		target.Data.ComboStaleCount != 2 ||
		target.Data.ComboManualReviewCount != 1 ||
		target.Data.OldestComboExceptionTime != 901 ||
		target.Data.ComboInvariantIssueCount != 3 ||
		target.Data.ComboIncompleteMatchGroupCount != 4 {
		t.Fatalf("combo operations watermarks=%+v", target.Data)
	}

	requestSource := &types.GetOperationsOverviewReq{
		TenantId: 9, RiskStaleSeconds: 30, ComboStaleSeconds: 45,
	}
	var requestTarget option.GetOperationsOverviewReq
	if err := copyValue(reflect.ValueOf(&requestTarget), reflect.ValueOf(requestSource)); err != nil {
		t.Fatal(err)
	}
	if requestTarget.TenantId != 9 ||
		requestTarget.RiskStaleSeconds != 30 ||
		requestTarget.ComboStaleSeconds != 45 {
		t.Fatalf("operations request tenant/risk/combo=%d/%d/%d",
			requestTarget.TenantId,
			requestTarget.RiskStaleSeconds,
			requestTarget.ComboStaleSeconds)
	}
}
