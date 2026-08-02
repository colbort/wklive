package models

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"wklive/proto/option"
)

func TestInsuranceFundSignedAmountUsesFlowTypeDirection(t *testing.T) {
	tests := []struct {
		name     string
		flowType option.InsuranceFundFlowType
		want     string
	}{
		{"liquidation fee inflow", option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_LIQUIDATION_FEE, "12.5"},
		{"deficit cover outflow", option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER, "-12.5"},
		{"manual deposit inflow", option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_MANUAL_DEPOSIT, "12.5"},
		{"manual withdraw outflow", option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_MANUAL_WITHDRAW, "-12.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := insuranceFundSignedAmount(int64(tt.flowType), decimal.RequireFromString("12.5"))
			if err != nil || !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("signed amount=%s err=%v want=%s", got, err, tt.want)
			}
		})
	}
	for _, invalid := range []struct {
		flowType int64
		amount   string
	}{
		{0, "1"},
		{int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER), "0"},
		{int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER), "-1"},
	} {
		if _, err := insuranceFundSignedAmount(invalid.flowType, decimal.RequireFromString(invalid.amount)); err == nil {
			t.Fatalf("invalid flow type/amount accepted: %+v", invalid)
		}
	}
	if optionInsuranceFundSignedAmountSQL !=
		"CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END" {
		t.Fatalf("unexpected signed SQL expression: %s", optionInsuranceFundSignedAmountSQL)
	}
}

func TestInsuranceFundFlowInsertRejectsInvalidEvidenceBeforeDatabase(t *testing.T) {
	model := &customTOptionInsuranceFundFlowModel{}
	valid := &TOptionInsuranceFundFlow{
		TenantId: 9, FlowNo: "INS-1", ContractId: 10, LiquidationId: 11,
		FlowType: int64(option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER),
		Coin:     "USDT", Amount: decimal.NewFromInt(3), AssetFlowNo: "ASSET-1", CreateTimes: 100,
	}
	tests := []struct {
		name   string
		mutate func(*TOptionInsuranceFundFlow)
		match  string
	}{
		{"negative magnitude", func(item *TOptionInsuranceFundFlow) { item.Amount = decimal.NewFromInt(-3) }, "positive magnitude"},
		{"missing asset evidence", func(item *TOptionInsuranceFundFlow) { item.AssetFlowNo = "" }, "identity"},
		{"missing liquidation", func(item *TOptionInsuranceFundFlow) { item.LiquidationId = 0 }, "requires contract and liquidation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := *valid
			tt.mutate(&item)
			if _, err := model.Insert(context.Background(), &item); err == nil || !strings.Contains(err.Error(), tt.match) {
				t.Fatalf("invalid insert err=%v want substring %q", err, tt.match)
			}
		})
	}
}
