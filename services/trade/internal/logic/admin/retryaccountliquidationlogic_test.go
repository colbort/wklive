package adminlogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestResetSettlementInstruction(t *testing.T) {
	instruction := &models.TTradeSettlementInstruction{
		Status:     int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW),
		RetryCount: 20, NextRetryAt: 99, LastErrorMsg: "failed", UpdateTimes: 1,
	}
	resetSettlementInstruction(instruction, 123)
	if instruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING) ||
		instruction.RetryCount != 0 || instruction.NextRetryAt != 123 ||
		instruction.LastErrorMsg != "" || instruction.UpdateTimes != 123 {
		t.Fatalf("unexpected reset instruction: %+v", instruction)
	}
}

func TestAccountLiquidationRetryStage(t *testing.T) {
	tests := []struct {
		name        string
		parent      *models.TContractAccountLiquidation
		hasADLFacts bool
		want        int64
	}{
		{
			name:   "positive account closes",
			parent: &models.TContractAccountLiquidation{},
			want:   models.ContractAccountLiquidationStatusClosing,
		},
		{
			name:   "insurance not started",
			parent: &models.TContractAccountLiquidation{DeficitAmount: decimal.NewFromInt(30)},
			want:   models.ContractAccountLiquidationStatusInsuranceFund,
		},
		{
			name: "partial insurance resumes ADL",
			parent: &models.TContractAccountLiquidation{
				DeficitAmount: decimal.NewFromInt(30), InsuranceFundAmount: decimal.NewFromInt(10),
			},
			want: models.ContractAccountLiquidationStatusADL,
		},
		{
			name:        "zero insurance with frozen ADL facts resumes ADL",
			parent:      &models.TContractAccountLiquidation{DeficitAmount: decimal.NewFromInt(30)},
			hasADLFacts: true,
			want:        models.ContractAccountLiquidationStatusADL,
		},
		{
			name: "covered deficit closes",
			parent: &models.TContractAccountLiquidation{
				DeficitAmount: decimal.NewFromInt(30), InsuranceFundAmount: decimal.NewFromInt(10),
				AdlReliefAmount: decimal.NewFromInt(20),
			},
			want: models.ContractAccountLiquidationStatusClosing,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accountLiquidationRetryStage(tt.parent, tt.hasADLFacts); got != tt.want {
				t.Fatalf("stage=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestRetryableSettlementInstructionStatus(t *testing.T) {
	if !isRetryableSettlementInstruction(int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)) ||
		!isRetryableSettlementInstruction(int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)) {
		t.Fatal("failed and manual-review instructions must be retryable")
	}
	if isRetryableSettlementInstruction(int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)) {
		t.Fatal("successful instruction must not be reset")
	}
}
