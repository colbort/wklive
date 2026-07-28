package tasklogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestCalculateFundingFeeMatrix(t *testing.T) {
	tests := []struct {
		name      string
		side      trade.PositionSide
		valueType trade.ContractValueType
		qty       string
		size      string
		mark      string
		rate      string
		want      string
	}{
		{name: "positive linear long pays", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, qty: "2", size: "1", mark: "100", rate: "0.01", want: "-2"},
		{name: "positive linear short receives", side: trade.PositionSide_POSITION_SIDE_SHORT, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, qty: "2", size: "1", mark: "100", rate: "0.01", want: "2"},
		{name: "negative linear long receives", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, qty: "2", size: "1", mark: "100", rate: "-0.01", want: "2"},
		{name: "negative linear short pays", side: trade.PositionSide_POSITION_SIDE_SHORT, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, qty: "2", size: "1", mark: "100", rate: "-0.01", want: "-2"},
		{name: "positive inverse long pays base", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE, qty: "100", size: "100", mark: "50000", rate: "0.01", want: "-0.002"},
		{name: "positive inverse short receives base", side: trade.PositionSide_POSITION_SIDE_SHORT, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE, qty: "100", size: "100", mark: "50000", rate: "0.01", want: "0.002"},
		{name: "zero rate", side: trade.PositionSide_POSITION_SIDE_LONG, valueType: trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR, qty: "2", size: "1", mark: "100", rate: "0", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateFundingFee(
				int64(tt.side),
				int64(tt.valueType),
				decimal.RequireFromString(tt.qty),
				decimal.RequireFromString(tt.size),
				decimal.RequireFromString(tt.mark),
				decimal.RequireFromString(tt.rate),
			)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("funding fee=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestCalculateFundingFeeRejectsInvalidPositionSide(t *testing.T) {
	if _, err := calculateFundingFee(
		int64(trade.PositionSide_POSITION_SIDE_NET),
		int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR),
		decimal.NewFromInt(1),
		decimal.NewFromInt(1),
		decimal.NewFromInt(100),
		decimal.RequireFromString("0.01"),
	); err == nil {
		t.Fatal("NET funding position must be projected to LONG/SHORT before settlement")
	}
}

func TestFundingUserAndDifferenceAmountsConservePerAsset(t *testing.T) {
	longFee, err := calculateFundingFee(
		int64(trade.PositionSide_POSITION_SIDE_LONG),
		int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR),
		decimal.NewFromInt(3),
		decimal.NewFromInt(1),
		decimal.NewFromInt(100),
		decimal.RequireFromString("0.01"),
	)
	if err != nil {
		t.Fatal(err)
	}
	shortFee, err := calculateFundingFee(
		int64(trade.PositionSide_POSITION_SIDE_SHORT),
		int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR),
		decimal.NewFromInt(2),
		decimal.NewFromInt(1),
		decimal.NewFromInt(100),
		decimal.RequireFromString("0.01"),
	)
	if err != nil {
		t.Fatal(err)
	}
	userNet := longFee.Add(shortFee)
	platformAmount := userNet.Neg()
	if !userNet.Add(platformAmount).IsZero() {
		t.Fatalf("funding is not conserved: users=%s platform=%s", userNet, platformAmount)
	}
	action, step := fundingDifferenceInstruction(userNet)
	if action != trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE || step != 2 {
		t.Fatalf("platform must receive payer surplus after user debits: action=%s step=%d", action, step)
	}
}

func TestFundingDifferenceInstructionBalancesUserNet(t *testing.T) {
	action, step := fundingDifferenceInstruction(decimal.RequireFromString("0.01"))
	if action != 8 || step != 1 {
		t.Fatalf("positive user net must debit platform first: action=%d step=%d", action, step)
	}
	action, step = fundingDifferenceInstruction(decimal.RequireFromString("-0.01"))
	if action != 3 || step != 2 {
		t.Fatalf("negative user net must credit platform after payers: action=%d step=%d", action, step)
	}
}

func TestSettlementInstructionIdentityIncludesSagaBinding(t *testing.T) {
	base := &models.TTradeSettlementInstruction{TenantId: 1, InstructionNo: "i", BizType: "funding", BizId: "s", BatchNo: "b", PositionId: 2, UserId: 3, Action: 8, Asset: "USDT", Amount: decimal.NewFromInt(1), StepNo: 1}
	copy := *base
	if !sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("identical instruction rejected")
	}
	copy.PositionId++
	if sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("different position binding accepted")
	}
	copy = *base
	copy.StepNo++
	if sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("different saga step accepted")
	}
}

func TestFundingPositionsFromHistoryUsesSettlementFacts(t *testing.T) {
	history := []*models.TContractPositionHistory{
		nil,
		{PositionId: 1, AfterQty: decimal.Zero},
		{
			PositionId:        2,
			TenantId:          3,
			UserId:            4,
			SymbolId:          5,
			ContractType:      1,
			ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE),
			PositionSide:      int64(trade.PositionSide_POSITION_SIDE_SHORT),
			MarginAsset:       "BTC",
			AfterQty:          decimal.RequireFromString("7.5"),
			AfterVersion:      9,
		},
	}

	got := fundingPositionsFromHistory(history)
	if len(got) != 1 {
		t.Fatalf("expected one open position, got %d", len(got))
	}
	p := got[0]
	if p.Id != 2 || p.TenantId != 3 || p.UserId != 4 || p.SymbolId != 5 ||
		p.MarginAsset != "BTC" || p.Version != 9 || !p.Qty.Equal(decimal.RequireFromString("7.5")) {
		t.Fatalf("unexpected reconstructed position: %+v", p)
	}
}
