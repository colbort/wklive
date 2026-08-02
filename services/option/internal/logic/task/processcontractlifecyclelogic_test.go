package tasklogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestSeriesContractLaunchApprovalAndControls(t *testing.T) {
	now := int64(1_000)
	series := &models.TOptionContractSeries{
		Status: int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED),
		LaunchStatus: int64(
			option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_PENDING_REVIEW,
		),
	}
	market := &models.TOptionMarket{
		UnderlyingPrice:        decimal.NewFromInt(100),
		MarkPrice:              decimal.NewFromInt(10),
		UnderlyingSnapshotTime: now,
		MarkSnapshotTime:       now,
		GreeksSnapshotTime:     now,
	}
	if seriesContractLaunchApproved(series) {
		t.Fatal("series contract must remain pending before independent launch approval")
	}
	series.LaunchStatus = int64(
		option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_APPROVED,
	)
	if !seriesContractLaunchApproved(series) {
		t.Fatal("generated series with independent approval should pass the series approval gate")
	}
	contract := &models.TOptionContract{
		Status:             int64(option.ContractStatus_CONTRACT_STATUS_PENDING),
		IsDeleted:          int64(common.YesNo_YES_NO_NO),
		ListTime:           now - 100,
		LastTradeTime:      now + 100,
		ExerciseCutoffTime: now + 200,
		ExpireTime:         now + 300,
		DeliverTime:        now + 400,
		MaxUserLongQty:     decimal.NewFromInt(100), MaxUserShortQty: decimal.NewFromInt(100),
		MaxOpenInterest:     decimal.NewFromInt(1000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"),
		CircuitBreakerRatio: decimal.RequireFromString("0.3"),
		GreeksMaxAgeSeconds: 60,
	}
	if !contractLaunchControlsReady(contract, market, now) {
		t.Fatal("complete controls and fresh market should pass runtime launch admission")
	}
	market.MarkSnapshotTime = now - 31
	if contractLaunchControlsReady(contract, market, now) {
		t.Fatal("contract must remain pending when mark price is stale")
	}
	market.MarkSnapshotTime = now
	market.GreeksSnapshotTime = now - 61
	if contractLaunchControlsReady(contract, market, now) {
		t.Fatal("contract must remain pending when Greeks exceed the approved threshold")
	}
	market.GreeksSnapshotTime = now
	contract.GreeksMaxAgeSeconds = 0
	if contractLaunchControlsReady(contract, market, now) {
		t.Fatal("contract must remain pending when Greeks threshold is unconfigured")
	}
	contract.GreeksMaxAgeSeconds = 60
	contract.MaxOpenInterest = decimal.Zero
	if contractLaunchControlsReady(contract, market, now) {
		t.Fatal("contract must remain pending when a mandatory risk control is zero")
	}
	series.Status = int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_REJECTED)
	if seriesContractLaunchApproved(series) {
		t.Fatal("non-generated series must never pass launch")
	}
}

func TestValidateOptionSettlementBalance(t *testing.T) {
	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(10),
	}); err != nil {
		t.Fatalf("balanced settlement rejected: %v", err)
	}

	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(9),
	}); err == nil {
		t.Fatal("unbalanced settlement must be rejected")
	}
}

func TestCalculateSettlementMedian(t *testing.T) {
	odd := []settlementPriceSample{
		{price: decimal.NewFromInt(110)},
		{price: decimal.NewFromInt(90)},
		{price: decimal.NewFromInt(100)},
	}
	if got := calculateSettlementMedian(odd); !got.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("odd median=%s want=100", got)
	}

	even := append(odd, settlementPriceSample{price: decimal.NewFromInt(120)})
	if got := calculateSettlementMedian(even); !got.Equal(decimal.NewFromInt(105)) {
		t.Fatalf("even median=%s want=105", got)
	}
}

func TestCalculateSettlementMedianIgnoresInvalidPrices(t *testing.T) {
	samples := []settlementPriceSample{
		{price: decimal.Zero},
		{price: decimal.NewFromInt(-1)},
		{price: decimal.NewFromInt(101)},
	}
	if got := calculateSettlementMedian(samples); !got.Equal(decimal.NewFromInt(101)) {
		t.Fatalf("median=%s want=101", got)
	}
}

func TestRejectedSettlementPriceIsNotRecreatedWithoutNewEvidence(t *testing.T) {
	latest := &models.TOptionSettlementPrice{
		PriceSource: "authoritative-market", WindowStart: 940, WindowEnd: 1000,
		SampleCount: 3, CalculationMethod: "MEDIAN", DeliveryPrice: decimal.NewFromInt(100),
		SourceSnapshotIds: `["a","b","c"]`,
		Status:            int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_REJECTED),
	}
	candidate := *latest
	candidate.Status = int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING)
	if !shouldSuppressRejectedSettlementPrice(latest, &candidate) {
		t.Fatal("identical rejected automatic evidence must not be recreated")
	}
	candidate.SourceSnapshotIds = `["a","b","d"]`
	if shouldSuppressRejectedSettlementPrice(latest, &candidate) {
		t.Fatal("new immutable evidence must be allowed to create a new review version")
	}
}

func TestApplyPositionSettlementReturnSplitsGrossAndFee(t *testing.T) {
	tests := []struct {
		name   string
		side   int64
		payoff string
		fee    string
		gross  string
		total  string
	}{
		{
			name: "long exercised", side: int64(common.PositionSide_POSITION_SIDE_LONG),
			payoff: "30", fee: "1", gross: "10", total: "9",
		},
		{
			name: "short assigned", side: int64(common.PositionSide_POSITION_SIDE_SHORT),
			payoff: "30", fee: "0", gross: "-10", total: "-10",
		},
		{
			name: "long abandoned", side: int64(common.PositionSide_POSITION_SIDE_LONG),
			payoff: "0", fee: "0", gross: "-20", total: "-20",
		},
		{
			name: "short expired", side: int64(common.PositionSide_POSITION_SIDE_SHORT),
			payoff: "0", fee: "0", gross: "20", total: "20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := &models.TOptionPosition{
				Side: tt.side, OpenAvgPrice: decimal.NewFromInt(5),
			}
			applyPositionSettlementReturn(
				position,
				decimal.RequireFromString(tt.payoff),
				decimal.RequireFromString(tt.fee),
				decimal.NewFromInt(2),
				decimal.NewFromInt(2),
			)
			if !position.SettlementRealizedPnl.Equal(decimal.RequireFromString(tt.gross)) ||
				!position.FeePaid.Equal(decimal.RequireFromString(tt.fee)) ||
				!position.TotalReturn.Equal(decimal.RequireFromString(tt.total)) ||
				!position.RealizedPnl.Equal(position.TotalReturn) {
				t.Fatalf(
					"split result gross=%s fee=%s total=%s realized=%s",
					position.SettlementRealizedPnl, position.FeePaid,
					position.TotalReturn, position.RealizedPnl,
				)
			}
		})
	}
}

func TestShouldExerciseAtExpiryHonorsThresholdAndInstruction(t *testing.T) {
	contract := &models.TOptionContract{AutoExerciseThreshold: decimal.NewFromInt(2)}
	intrinsic := decimal.NewFromInt(1)
	payoff := decimal.NewFromInt(10)
	fee := decimal.NewFromInt(1)
	if shouldExerciseAtExpiry(
		contract, intrinsic, payoff, fee,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO,
	) {
		t.Fatal("auto exercise below threshold must be skipped")
	}
	if !shouldExerciseAtExpiry(
		contract, intrinsic, payoff, fee,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE,
	) {
		t.Fatal("contrary exercise with positive net payoff must be accepted")
	}
	if shouldExerciseAtExpiry(
		contract, decimal.NewFromInt(3), payoff, fee,
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE,
	) {
		t.Fatal("DNE must override automatic exercise")
	}
	if shouldExerciseAtExpiry(
		contract, decimal.NewFromInt(3), decimal.NewFromInt(1), decimal.NewFromInt(1),
		option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE,
	) {
		t.Fatal("zero net payoff must not exercise")
	}
}

func TestAllocateExpiryShortQuantityFIFOAndDNEBalance(t *testing.T) {
	positions := []*models.TOptionPosition{
		{Id: 1, Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(2), Status: int64(option.PositionStatus_POSITION_STATUS_EXERCISED)},
		{Id: 2, Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(3), Status: int64(option.PositionStatus_POSITION_STATUS_EXPIRED)},
		{Id: 11, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1), CreateTimes: 10},
		{Id: 12, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(4), CreateTimes: 20},
	}
	allocated, err := allocateExpiryShortQuantity(positions)
	if err != nil {
		t.Fatalf("allocate expiry shorts: %v", err)
	}
	if !allocated[11].Equal(decimal.NewFromInt(1)) ||
		!allocated[12].Equal(decimal.NewFromInt(1)) {
		t.Fatalf("unexpected FIFO allocation: %#v", allocated)
	}
}

func TestPhysicalLongAssetLegs(t *testing.T) {
	contract := &models.TOptionContract{
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		UnderlyingCoin: "BTC", SettleCoin: "USDT",
	}
	debitCoin, debit, creditCoin, credit := physicalLongAssetLegs(
		contract, decimal.NewFromInt(2), decimal.NewFromInt(200),
	)
	if debitCoin != "USDT" || !debit.Equal(decimal.NewFromInt(200)) ||
		creditCoin != "BTC" || !credit.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("unexpected call delivery legs: %s %s -> %s %s", debitCoin, debit, creditCoin, credit)
	}
	contract.OptionType = int64(option.OptionType_OPTION_TYPE_PUT)
	debitCoin, debit, creditCoin, credit = physicalLongAssetLegs(
		contract, decimal.NewFromInt(2), decimal.NewFromInt(200),
	)
	if debitCoin != "BTC" || !debit.Equal(decimal.NewFromInt(2)) ||
		creditCoin != "USDT" || !credit.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("unexpected put delivery legs: %s %s -> %s %s", debitCoin, debit, creditCoin, credit)
	}
}

func TestAllocatePhysicalDeliveryPositionsPairsFIFOWithoutCrossUnitLeakage(t *testing.T) {
	positions := []*models.TOptionPosition{
		{Id: 1, Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(3)},
		{Id: 2, Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(2)},
		{Id: 11, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1)},
		{Id: 12, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(4)},
	}
	units, err := allocatePhysicalDeliveryPositions(positions)
	if err != nil {
		t.Fatalf("allocate physical delivery: %v", err)
	}
	if len(units) != 3 {
		t.Fatalf("units=%d want=3", len(units))
	}
	expected := []struct {
		longID, shortID int64
		quantity        string
	}{
		{1, 11, "1"},
		{1, 12, "2"},
		{2, 12, "2"},
	}
	for i, unit := range units {
		if unit.long.Id != expected[i].longID || unit.short.Id != expected[i].shortID ||
			!unit.quantity.Equal(decimal.RequireFromString(expected[i].quantity)) {
			t.Fatalf(
				"unit[%d]=long:%d short:%d qty:%s",
				i, unit.long.Id, unit.short.Id, unit.quantity,
			)
		}
	}
}

func TestAllocatePhysicalDeliveryPositionsRejectsUnbalancedBook(t *testing.T) {
	_, err := allocatePhysicalDeliveryPositions([]*models.TOptionPosition{
		{Id: 1, Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(2)},
		{Id: 11, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1)},
	})
	if err == nil {
		t.Fatal("unbalanced physical book must be rejected before asset instructions")
	}
}
