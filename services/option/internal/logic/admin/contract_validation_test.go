package adminlogic

import (
	"context"
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestCreateContractRejectsDirectTradingStatus(t *testing.T) {
	logic := NewCreateContractLogic(context.Background(), nil)
	resp, err := logic.CreateContract(&option.CreateContractReq{
		Status: option.ContractStatus_CONTRACT_STATUS_TRADING,
	})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code == 0 {
		t.Fatalf("direct TRADING creation must be rejected before persistence: resp=%+v err=%v", resp, err)
	}
}

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
		LiquidationDeficitPolicy: int64(
			option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
		),
		Status: int64(option.ContractStatus_CONTRACT_STATUS_PENDING),
	}
}

func TestValidateSupportedContractDeficitPolicy(t *testing.T) {
	item := validContractForTest()
	item.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED)
	item.InitialMarginRate = decimal.RequireFromString("0.2")
	item.MaintenanceMarginRate = decimal.RequireFromString("0.1")
	item.MinMarginRate = decimal.RequireFromString("0.05")
	item.InsuranceUserId = 10
	item.InsuranceAccountId = 20
	item.LiquidationDeficitPolicy = int64(
		option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP,
	)
	if !validateSupportedContract(item) {
		t.Fatal("isolated contract with platform backstop should be valid")
	}
	item.LiquidationDeficitPolicy = int64(
		option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_UNKNOWN,
	)
	if validateSupportedContract(item) {
		t.Fatal("unknown liquidation deficit policy must be rejected")
	}
}

func TestValidateSupportedPortfolioContract(t *testing.T) {
	item := validContractForTest()
	item.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
	item.InitialMarginRate = decimal.RequireFromString("0.2")
	item.MaintenanceMarginRate = decimal.RequireFromString("0.1")
	item.MinMarginRate = decimal.RequireFromString("0.05")
	item.InsuranceUserId = 10
	item.InsuranceAccountId = 20
	if !validateSupportedContract(item) {
		t.Fatal("portfolio contract with complete risk parameters should be supported")
	}
}

func TestValidateSupportedContract(t *testing.T) {
	item := validContractForTest()
	if !validateSupportedContract(item) {
		t.Fatal("expected valid restricted contract")
	}

	physical := *item
	physical.SettlementType = int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL)
	physical.UnderlyingCoin = "BTC"
	physical.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY)
	physical.PhysicalDeliveryPolicy = int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT)
	physical.PhysicalDeliveryCureSeconds = 3600
	if !validateSupportedContract(&physical) {
		t.Fatal("strict fully covered physical settlement should be supported")
	}
	physical.UnderlyingCoin = ""
	if validateSupportedContract(&physical) {
		t.Fatal("physical settlement without an underlying coin must be rejected")
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

	invalidCutoff := *item
	invalidCutoff.ExerciseCutoffTime = invalidCutoff.ListTime
	if validateSupportedContract(&invalidCutoff) {
		t.Fatal("exercise cutoff must be after list time")
	}

	invalidThreshold := *item
	invalidThreshold.AutoExerciseThreshold = decimal.NewFromInt(-1)
	if validateSupportedContract(&invalidThreshold) {
		t.Fatal("negative auto exercise threshold must be rejected")
	}

	physicalThreshold := physical
	physicalThreshold.UnderlyingCoin = "BTC"
	physicalThreshold.AutoExerciseThreshold = decimal.NewFromInt(1)
	if validateSupportedContract(&physicalThreshold) {
		t.Fatal("physical delivery does not yet support threshold/DNE allocation")
	}
	physicalCure := physical
	physicalCure.AutoExerciseThreshold = decimal.Zero
	physicalCure.PhysicalDeliveryCureSeconds = minPhysicalCureSeconds - 1
	if validateSupportedContract(&physicalCure) {
		t.Fatal("physical delivery cure period below minimum must be rejected")
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
	right.GreeksMaxAgeSeconds = 60
	if !economicContractFieldsEqual(left, &right) {
		t.Fatal("approved Greeks freshness is an editable audited control, not contract economics")
	}

	right.StrikePrice = right.StrikePrice.Add(decimal.NewFromInt(1))
	if economicContractFieldsEqual(left, &right) {
		t.Fatal("strike price is an immutable economic field")
	}
	right = *left
	right.TradingCalendarCode = "US_OPTIONS"
	if economicContractFieldsEqual(left, &right) {
		t.Fatal("trading calendar code is immutable after listing")
	}
}

func TestValidateSupportedContractTradingCalendarCode(t *testing.T) {
	item := validContractForTest()
	if !validateSupportedContract(item) || item.TradingCalendarCode != "CONTINUOUS_24_7" {
		t.Fatal("blank legacy calendar code should normalize to the compatibility calendar")
	}
	item.TradingCalendarCode = " us_options "
	if !validateSupportedContract(item) || item.TradingCalendarCode != "US_OPTIONS" {
		t.Fatal("valid calendar code should be normalized")
	}
	item.TradingCalendarCode = "bad/calendar"
	if validateSupportedContract(item) {
		t.Fatal("invalid trading calendar code accepted")
	}
}

func TestTradingContractRequiresExplicitControls(t *testing.T) {
	item := validContractForTest()
	item.Status = int64(option.ContractStatus_CONTRACT_STATUS_TRADING)
	if validateSupportedContract(item) {
		t.Fatal("trading contract with zero controls must be rejected")
	}
	item.MaxUserLongQty = decimal.NewFromInt(100)
	item.MaxUserShortQty = decimal.NewFromInt(50)
	item.MaxOpenInterest = decimal.NewFromInt(1000)
	item.OrderPriceBandRatio = decimal.RequireFromString("0.20")
	item.CircuitBreakerRatio = decimal.RequireFromString("0.30")
	item.GreeksMaxAgeSeconds = 60
	if !validateSupportedContract(item) {
		t.Fatal("trading contract with complete controls should be accepted")
	}
	item.GreeksMaxAgeSeconds = 0
	if validateSupportedContract(item) {
		t.Fatal("trading contract without an approved Greeks threshold must be rejected")
	}
	item.OrderPriceBandRatio = decimal.RequireFromString("1.01")
	if validateSupportedContract(item) {
		t.Fatal("price band ratio above one must be rejected")
	}
}

func TestPendingContractRejectsNegativeGreeksThreshold(t *testing.T) {
	item := validContractForTest()
	item.GreeksMaxAgeSeconds = -1
	if validateSupportedContract(item) {
		t.Fatal("negative Greeks threshold must be rejected in every contract state")
	}
}

func TestOptionCircuitBreakerDecision(t *testing.T) {
	ratio, trip := optionCircuitBreakerDecision(
		decimal.NewFromInt(100), decimal.NewFromInt(120), decimal.RequireFromString("0.20"),
	)
	if !trip || !ratio.Equal(decimal.RequireFromString("0.20")) {
		t.Fatalf("boundary must trip, ratio=%s trip=%t", ratio, trip)
	}
	_, trip = optionCircuitBreakerDecision(
		decimal.NewFromInt(100), decimal.NewFromInt(119), decimal.RequireFromString("0.20"),
	)
	if trip {
		t.Fatal("move below threshold must not trip")
	}
	_, trip = optionCircuitBreakerDecision(
		decimal.NewFromInt(100), decimal.NewFromInt(150), decimal.Zero,
	)
	if trip {
		t.Fatal("unconfigured circuit breaker must not trip silently")
	}
}

func TestMergeOptionMarketPatchPreservesConcurrentFields(t *testing.T) {
	target := &models.TOptionMarket{
		UnderlyingPrice:        decimal.NewFromInt(100),
		MarkPrice:              decimal.NewFromInt(10),
		BidPrice:               decimal.NewFromInt(9),
		Iv:                     decimal.RequireFromString("0.50"),
		UnderlyingSnapshotTime: 1000,
		MarkSnapshotTime:       1000,
		GreeksSnapshotTime:     1000,
		SnapshotTime:           1000,
		UpdateTimes:            1000,
	}
	patch := &models.TOptionMarket{
		MarkPrice:        decimal.NewFromInt(12),
		MarkSnapshotTime: 2000,
		SnapshotTime:     2000,
		UpdateTimes:      2001,
	}
	mergeOptionMarketPatch(target, patch, &option.UpdateMarketReq{MarkPrice: "12"})

	if !target.MarkPrice.Equal(decimal.NewFromInt(12)) || target.MarkSnapshotTime != 2000 {
		t.Fatalf("mark patch not applied: %+v", target)
	}
	if !target.UnderlyingPrice.Equal(decimal.NewFromInt(100)) ||
		!target.BidPrice.Equal(decimal.NewFromInt(9)) ||
		!target.Iv.Equal(decimal.RequireFromString("0.50")) {
		t.Fatalf("unrelated fields were overwritten: %+v", target)
	}
	if target.UnderlyingSnapshotTime != 1000 || target.GreeksSnapshotTime != 1000 {
		t.Fatalf("unrelated freshness timestamps were overwritten: %+v", target)
	}
}
