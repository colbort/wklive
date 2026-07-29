package tasklogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestCalculateCrossTakeoverSettlementCredit(t *testing.T) {
	fee, credit, debit, deficit := calculateCrossTakeoverSettlement(
		decimal.NewFromInt(30),
		decimal.NewFromInt(130),
		decimal.NewFromInt(3),
		decimal.NewFromInt(100),
	)
	if !fee.Equal(decimal.NewFromInt(3)) ||
		!credit.Equal(decimal.NewFromInt(27)) ||
		!debit.IsZero() || !deficit.IsZero() {
		t.Fatalf("unexpected credit settlement fee=%s credit=%s debit=%s deficit=%s", fee, credit, debit, deficit)
	}
}

func TestCalculateCrossTakeoverSettlementUsesSharedWallet(t *testing.T) {
	fee, credit, debit, deficit := calculateCrossTakeoverSettlement(
		decimal.NewFromInt(-80),
		decimal.NewFromInt(20),
		decimal.NewFromInt(5),
		decimal.NewFromInt(100),
	)
	if !fee.Equal(decimal.NewFromInt(5)) || !credit.IsZero() ||
		!debit.Equal(decimal.NewFromInt(85)) || !deficit.IsZero() {
		t.Fatalf("unexpected shared-wallet settlement fee=%s credit=%s debit=%s deficit=%s", fee, credit, debit, deficit)
	}
}

func TestCalculateCrossTakeoverSettlementCapsAtWalletAndReportsDeficit(t *testing.T) {
	fee, credit, debit, deficit := calculateCrossTakeoverSettlement(
		decimal.NewFromInt(-130),
		decimal.NewFromInt(-30),
		decimal.NewFromInt(5),
		decimal.NewFromInt(100),
	)
	if !fee.IsZero() || !credit.IsZero() ||
		!debit.Equal(decimal.NewFromInt(100)) || !deficit.Equal(decimal.NewFromInt(30)) {
		t.Fatalf("unexpected deficit settlement fee=%s credit=%s debit=%s deficit=%s", fee, credit, debit, deficit)
	}
}

func TestAllocateCrossLiquidationFeesPreservesTotal(t *testing.T) {
	fees := allocateCrossLiquidationFees(
		[]decimal.Decimal{decimal.NewFromInt(2), decimal.NewFromInt(3), decimal.NewFromInt(5)},
		decimal.NewFromInt(7),
	)
	total := decimal.Zero
	for _, fee := range fees {
		total = total.Add(fee)
	}
	if len(fees) != 3 || !total.Equal(decimal.NewFromInt(7)) {
		t.Fatalf("fee allocation=%v total=%s", fees, total)
	}
}

func TestCrossAccountProductionGateOnlyHoldsNewTakeover(t *testing.T) {
	if !shouldHoldCrossAccountLiquidation(false, models.ContractAccountLiquidationStatusPending) {
		t.Fatal("disabled gate must hold a new account takeover")
	}
	for _, status := range []int64{
		models.ContractAccountLiquidationStatusAssetSettling,
		models.ContractAccountLiquidationStatusInsuranceFund,
		models.ContractAccountLiquidationStatusADL,
		models.ContractAccountLiquidationStatusClosing,
	} {
		if shouldHoldCrossAccountLiquidation(false, status) {
			t.Fatalf("disabled gate must not strand started Saga status=%d", status)
		}
	}
}

func TestAllocateCrossADLDeficitsUsesFrozenLosses(t *testing.T) {
	items := []*models.TContractAccountLiquidationItem{
		{RealizedPnl: decimal.NewFromInt(-30)},
		{RealizedPnl: decimal.NewFromInt(5)},
		{RealizedPnl: decimal.NewFromInt(-20)},
	}
	got, err := allocateCrossADLDeficits(items, decimal.NewFromInt(25))
	if err != nil {
		t.Fatal(err)
	}
	if !got[0].Equal(decimal.NewFromInt(15)) || !got[1].IsZero() ||
		!got[2].Equal(decimal.NewFromInt(10)) {
		t.Fatalf("unexpected ADL targets: %v", got)
	}
	if _, err = allocateCrossADLDeficits(items, decimal.NewFromInt(51)); err == nil {
		t.Fatal("deficit larger than frozen losses must be rejected")
	}
}

func TestCrossAccountBankruptcyPriceLinear(t *testing.T) {
	tests := []struct {
		name string
		side trade.PositionSide
		mark string
		want string
	}{
		{name: "long", side: trade.PositionSide_POSITION_SIDE_LONG, mark: "70", want: "90"},
		{name: "short", side: trade.PositionSide_POSITION_SIDE_SHORT, mark: "130", want: "110"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := &models.TContractPosition{
				PositionSide: int64(tt.side), ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR),
				Qty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(100),
				MarkPrice: decimal.RequireFromString(tt.mark),
			}
			got, err := crossAccountBankruptcyPrice(position, decimal.NewFromInt(1), decimal.NewFromInt(20))
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("price=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestCrossAccountBankruptcyPriceInverse(t *testing.T) {
	tests := []struct {
		name   string
		side   trade.PositionSide
		mark   string
		relief string
	}{
		{name: "long", side: trade.PositionSide_POSITION_SIDE_LONG, mark: "50", relief: "0.5"},
		{name: "short", side: trade.PositionSide_POSITION_SIDE_SHORT, mark: "200", relief: "0.25"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := &models.TContractPosition{
				PositionSide: int64(tt.side), ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE),
				Qty: decimal.NewFromInt(100), OpenAvgPrice: decimal.NewFromInt(100),
				MarkPrice: decimal.RequireFromString(tt.mark),
			}
			relief := decimal.RequireFromString(tt.relief)
			got, err := crossAccountBankruptcyPrice(position, decimal.NewFromInt(1), relief)
			if err != nil {
				t.Fatal(err)
			}
			markPnl := contractRealizedPnl(position.PositionSide, position.OpenAvgPrice, position.MarkPrice, position.Qty, decimal.NewFromInt(1), position.ContractValueType)
			bankruptcyPnl := contractRealizedPnl(position.PositionSide, position.OpenAvgPrice, got, position.Qty, decimal.NewFromInt(1), position.ContractValueType)
			if !crossAmountCovered(relief, bankruptcyPnl.Sub(markPnl)) {
				t.Fatalf("price=%s only relieves %s want=%s", got, bankruptcyPnl.Sub(markPnl), relief)
			}
		})
	}
}

func TestCrossLiquidationFeeInstructionIsDurableAndPreReconciledOnSuccess(t *testing.T) {
	batch := &models.TContractAccountLiquidation{
		TenantId: 9, UserId: 8, LiquidationNo: "XLIQ-1-2", MarginAsset: "USDT",
	}
	instruction := crossLiquidationFeeInstruction(batch, decimal.NewFromInt(3), 100)
	if instruction.InstructionNo != "XLIQ-1-2-FEE" ||
		instruction.BizType != crossAccountLiquidationBizType ||
		instruction.Action != int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE) ||
		instruction.StepNo != 2 ||
		!instruction.Amount.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("unexpected fee instruction: %+v", instruction)
	}
}
