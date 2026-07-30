package adminlogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func validContractForTest() *models.TOptionContract {
	return &models.TOptionContract{
		TenantId: 1, ContractCode: "BTC-C-100", UnderlyingSymbol: "BTCUSDT",
		SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle:  int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.NewFromInt(1), MaxOrderQty: decimal.NewFromInt(100),
		PriceTick: decimal.RequireFromString("0.01"), QtyStep: decimal.NewFromInt(1),
		Multiplier: decimal.NewFromInt(1), ListTime: 100, ExpireTime: 200, DeliverTime: 210,
		IsAutoExercise:   int64(common.YesNo_YES_NO_YES),
		SellerMarginMode: int64(option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED),
		Status:           int64(option.ContractStatus_CONTRACT_STATUS_PENDING),
	}
}

func TestValidateSupportedContract(t *testing.T) {
	item := validContractForTest()
	if !validateSupportedContract(item) {
		t.Fatal("expected valid restricted contract")
	}

	physical := *item
	physical.SettlementType = int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL)
	if validateSupportedContract(&physical) {
		t.Fatal("physical settlement must be rejected before delivery support exists")
	}

	american := *item
	american.ExerciseStyle = int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN)
	if !validateSupportedContract(&american) {
		t.Fatal("cash-settled American exercise should be supported")
	}

	invalidTimes := *item
	invalidTimes.ExpireTime = invalidTimes.ListTime
	if validateSupportedContract(&invalidTimes) {
		t.Fatal("expire time must be later than list time")
	}
}

func TestValidateSupportedContractFeeAccount(t *testing.T) {
	item := validContractForTest()
	item.MakerFeeRate = decimal.RequireFromString("0.001")
	if validateSupportedContract(item) {
		t.Fatal("positive fee rate without collection account must be rejected")
	}
	item.FeeUserId = 100
	item.FeeAccountId = 200
	if !validateSupportedContract(item) {
		t.Fatal("valid fee configuration rejected")
	}
	item.TakerFeeRate = decimal.RequireFromString("1.01")
	if validateSupportedContract(item) {
		t.Fatal("fee rate above one must be rejected")
	}
}

func TestEconomicContractFieldsEqual(t *testing.T) {
	left := validContractForTest()
	right := *left
	right.Remark = "operational change"
	right.Sort = 10
	right.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
	if !economicContractFieldsEqual(left, &right) {
		t.Fatal("operational fields should remain editable")
	}

	right.StrikePrice = right.StrikePrice.Add(decimal.NewFromInt(1))
	if economicContractFieldsEqual(left, &right) {
		t.Fatal("strike price is an immutable economic field")
	}
}
