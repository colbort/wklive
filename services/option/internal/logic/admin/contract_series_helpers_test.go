package adminlogic

import (
	"testing"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func validContractSeriesRequest() *option.CreateContractSeriesReq {
	now := time.Now().Unix()
	return &option.CreateContractSeriesReq{
		TenantId: 1, RequestKey: "series-request-001", SeriesCode: "BTCUSD",
		ReferencePrice: "100", ReferenceSource: "index-snapshot",
		ReferenceTime: now, EvidenceRef: "evidence://series/001", ChangeReason: "initial series",
		ContractTemplate: &option.CreateContractReq{
			UnderlyingSymbol: "BTCUSDT", SettleCoin: "USDT", QuoteCoin: "USDT",
			ExerciseStyle:  option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
			SettlementType: option.SettlementType_SETTLEMENT_TYPE_CASH,
			ContractUnit:   "1", MinOrderQty: "1", MaxOrderQty: "100",
			PriceTick: "0.01", QtyStep: "1", Multiplier: "1",
			IsAutoExercise:           common.YesNo_YES_NO_YES,
			SellerMarginMode:         option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED,
			LiquidationDeficitPolicy: option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
			AutoExerciseThreshold:    "0", MaxUserLongQty: "100",
			MaxUserShortQty: "100", MaxOpenInterest: "1000",
			OrderPriceBandRatio: "0.2", CircuitBreakerRatio: "0.3",
			TradingCalendarCode: "CONTINUOUS_24_7",
		},
		Expiries: []*option.ContractSeriesExpiryInput{
			{SequenceNo: 1, CycleCode: "WEEKLY", ListTime: now + 60,
				ExerciseCutoffTime: now + 3600, ExpireTime: now + 3600, DeliverTime: now + 3660},
			{SequenceNo: 2, CycleCode: "MONTHLY", ListTime: now + 60,
				ExerciseCutoffTime: now + 7200, ExpireTime: now + 7200, DeliverTime: now + 7260},
		},
		StrikeBands: []*option.ContractSeriesStrikeBandInput{
			{SequenceNo: 1, LowerStrike: "80", UpperStrike: "120", StrikeStep: "20"},
		},
	}
}

func TestPrepareContractSeriesDeterministicAndSymmetric(t *testing.T) {
	in := validContractSeriesRequest()
	prepared, err := prepareContractSeries(in)
	if err != nil {
		t.Fatalf("prepare series: %v", err)
	}
	if prepared.expectedCount != 12 || len(prepared.strikes) != 3 {
		t.Fatalf("unexpected ladder: strikes=%d contracts=%d", len(prepared.strikes), prepared.expectedCount)
	}
	if prepared.template.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
		prepared.template.OptionType != int64(option.OptionType_OPTION_TYPE_CALL) ||
		!prepared.template.StrikePrice.Equal(decimal.NewFromInt(100)) {
		t.Fatal("template overrides were not applied")
	}
	firstHash := prepared.payloadHash
	in.ContractTemplate.ContractCode = "IGNORED"
	in.ContractTemplate.StrikePrice = "999999"
	in.ContractTemplate.Status = option.ContractStatus_CONTRACT_STATUS_TRADING
	in.ContractTemplate.ListTime = 1
	second, err := prepareContractSeries(in)
	if err != nil {
		t.Fatalf("prepare replay: %v", err)
	}
	if second.payloadHash != firstHash {
		t.Fatal("ignored contract identity fields must not change the idempotency payload")
	}
}

func TestPrepareContractSeriesRejectsOverlappingAndInexactBands(t *testing.T) {
	in := validContractSeriesRequest()
	in.StrikeBands = append(in.StrikeBands, &option.ContractSeriesStrikeBandInput{
		SequenceNo: 2, LowerStrike: "120", UpperStrike: "140", StrikeStep: "10",
	})
	if _, err := prepareContractSeries(in); err == nil {
		t.Fatal("overlapping closed strike bands must be rejected")
	}
	in = validContractSeriesRequest()
	in.StrikeBands[0].UpperStrike = "121"
	if _, err := prepareContractSeries(in); err == nil {
		t.Fatal("band whose upper bound is not exactly reachable must be rejected")
	}
	in = validContractSeriesRequest()
	in.StrikeBands[0].StrikeStep = "0.00000000000000001"
	if _, err := prepareContractSeries(in); err == nil {
		t.Fatal("scale above DECIMAL(32,16) must be rejected")
	}
}

func TestContractSeriesFiveHundredBoundary(t *testing.T) {
	in := validContractSeriesRequest()
	in.Expiries = in.Expiries[:1]
	in.StrikeBands[0] = &option.ContractSeriesStrikeBandInput{
		SequenceNo: 1, LowerStrike: "1", UpperStrike: "250", StrikeStep: "1",
	}
	prepared, err := prepareContractSeries(in)
	if err != nil {
		t.Fatalf("500-contract boundary rejected: %v", err)
	}
	if prepared.expectedCount != 500 {
		t.Fatalf("expected 500 contracts, got %d", prepared.expectedCount)
	}
	in.StrikeBands[0].UpperStrike = "251"
	if _, err := prepareContractSeries(in); err == nil {
		t.Fatal("501+ contract request must fail before database writes")
	}
}

func TestCloneSeriesContractForcesPendingAndLineage(t *testing.T) {
	prepared, err := prepareContractSeries(validContractSeriesRequest())
	if err != nil {
		t.Fatal(err)
	}
	series := &models.TOptionContractSeries{
		Id: 99, TenantId: 1, SeriesCode: "BTCUSD", Version: 2,
	}
	expiry := prepared.expiries[0]
	expiry.Id = 10
	item, err := cloneSeriesContract(
		prepared.template, series, expiry, prepared.strikes[0], 0,
		option.OptionType_OPTION_TYPE_PUT, time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("clone contract: %v", err)
	}
	if item.ContractCode != "BTCUSD-V2-E001-K001-P" ||
		item.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
		item.OptionType != int64(option.OptionType_OPTION_TYPE_PUT) ||
		item.ExpireTime != expiry.ExpireTime {
		t.Fatalf("unexpected generated contract: %+v", item)
	}
}
