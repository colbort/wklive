package mapper

import (
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func RiskTierProto(v *models.TContractRiskLimitTier) *trade.ContractRiskLimitTier {
	return &trade.ContractRiskLimitTier{Id: v.Id, TenantId: v.TenantId, SymbolId: v.SymbolId, TierNo: v.TierNo, NotionalFloor: v.NotionalFloor.String(), NotionalCap: v.NotionalCap.String(), MaxLeverage: v.MaxLeverage, InitialMarginRate: v.InitialMarginRate.String(), MaintenanceMarginRate: v.MaintenanceMarginRate.String(), MaintenanceAmount: v.MaintenanceAmount.String(), Enabled: common.Enable(v.Enabled), CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}
func FundingBatchProto(v *models.TContractFundingBatch) *trade.ContractFundingBatch {
	return &trade.ContractFundingBatch{Id: v.Id, BatchNo: v.BatchNo, SymbolId: v.SymbolId, FundingRate: v.FundingRate.String(), MarkPrice: v.MarkPrice.String(), IndexPrice: v.IndexPrice.String(), PriceSource: v.PriceSource, FormulaVersion: v.FormulaVersion, SettlementTime: v.SettlementTime, Status: trade.FundingBatchStatus(v.Status), TotalPositions: v.TotalPositions, SettledPositions: v.SettledPositions, LastErrorMsg: v.LastErrorMsg, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}
func FundingSettlementProto(v *models.TContractFundingSettlement) *trade.ContractFundingSettlement {
	return &trade.ContractFundingSettlement{Id: v.Id, SettlementNo: v.SettlementNo, BatchId: v.BatchId, BatchNo: v.BatchNo, SymbolId: v.SymbolId, UserId: v.UserId, PositionId: v.PositionId, PositionSide: trade.PositionSide(v.PositionSide), FundingRate: v.FundingRate.String(), MarkPrice: v.MarkPrice.String(), PositionQty: v.PositionQty.String(), FeeAsset: v.FeeAsset, FeeAmount: v.FeeAmount.String(), SettlementTime: v.SettlementTime, Status: trade.FundingSettlementStatus(v.Status), RetryCount: v.RetryCount, NextRetryAt: v.NextRetryAt, LastErrorMsg: v.LastErrorMsg, SettledAt: v.SettledAt, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes, PositionVersion: v.PositionVersion}
}
func DeliveryBatchProto(v *models.TContractDeliveryBatch) *trade.ContractDeliveryBatch {
	sampleSnapshot := ""
	if v.SampleSnapshot.Valid {
		sampleSnapshot = v.SampleSnapshot.String
	}
	return &trade.ContractDeliveryBatch{Id: v.Id, BatchNo: v.BatchNo, SymbolId: v.SymbolId, SettlementPrice: v.SettlementPrice.String(), PriceSource: v.PriceSource, PriceAlgorithm: v.PriceAlgorithm, OpenCutoffTime: v.OpenCutoffTime, MatchingStopTime: v.MatchingStopTime, DeliveryTime: v.DeliveryTime, Status: trade.DeliveryBatchStatus(v.Status), TotalPositions: v.TotalPositions, SettledPositions: v.SettledPositions, LastErrorMsg: v.LastErrorMsg, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes, FormulaVersion: v.FormulaVersion, SampleSnapshot: sampleSnapshot}
}
func DeliverySettlementProto(v *models.TContractDeliverySettlement) *trade.ContractDeliverySettlement {
	return &trade.ContractDeliverySettlement{Id: v.Id, SettlementNo: v.SettlementNo, BatchId: v.BatchId, BatchNo: v.BatchNo, SymbolId: v.SymbolId, UserId: v.UserId, PositionId: v.PositionId, PositionSide: trade.PositionSide(v.PositionSide), SettlementPrice: v.SettlementPrice.String(), PositionQty: v.PositionQty.String(), RealizedPnl: v.RealizedPnl.String(), DeliveryFee: v.DeliveryFee.String(), SettleAsset: v.SettleAsset, DeliveryTime: v.DeliveryTime, Status: trade.DeliverySettlementStatus(v.Status), RetryCount: v.RetryCount, NextRetryAt: v.NextRetryAt, LastErrorMsg: v.LastErrorMsg, SettledAt: v.SettledAt, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}
func LiquidationProto(v *models.TContractLiquidation) *trade.ContractLiquidation {
	return &trade.ContractLiquidation{Id: v.Id, LiquidationNo: v.LiquidationNo, PositionId: v.PositionId, UserId: v.UserId, SymbolId: v.SymbolId, PositionSide: trade.PositionSide(v.PositionSide), MarginMode: trade.MarginMode(v.MarginMode), TriggerMarkPrice: v.TriggerMarkPrice.String(), TriggerIndexPrice: v.TriggerIndexPrice.String(), TriggerSnapshotId: v.TriggerSnapshotId, TriggerQty: v.TriggerQty.String(), LiquidatedQty: v.LiquidatedQty.String(), MaintenanceMargin: v.MaintenanceMargin.String(), AccountEquity: v.AccountEquity.String(), BankruptcyPrice: v.BankruptcyPrice.String(), LiquidationFee: v.LiquidationFee.String(), InsuranceFundAmount: v.InsuranceFundAmount.String(), AdlQty: v.AdlQty.String(), Status: trade.LiquidationStatus(v.Status), Reason: v.Reason, StartedAt: v.StartedAt, CompletedAt: v.CompletedAt, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}
func AccountLiquidationProto(v *models.TContractAccountLiquidation) *trade.ContractAccountLiquidation {
	return &trade.ContractAccountLiquidation{
		Id: v.Id, LiquidationNo: v.LiquidationNo, UserId: v.UserId, MarginAsset: v.MarginAsset,
		MarginSnapshotId: v.MarginSnapshotId, MarginSnapshotVersion: v.MarginSnapshotVersion,
		AssetVersion: v.AssetVersion, WalletBalance: v.WalletBalance.String(),
		PositionMargin: v.PositionMargin.String(), MaintenanceMargin: v.MaintenanceMargin.String(),
		AccountEquity: v.AccountEquity.String(), RiskRate: v.RiskRate.String(),
		GrossSettlement: v.GrossSettlement.String(), LiquidationFee: v.LiquidationFee.String(),
		UserCredit: v.UserCredit.String(), UserDebit: v.UserDebit.String(),
		DeficitAmount: v.DeficitAmount.String(), InsuranceFundAmount: v.InsuranceFundAmount.String(),
		AdlReliefAmount: v.AdlReliefAmount.String(), AdlQty: v.AdlQty.String(),
		PositionCount: v.PositionCount,
		Status:        trade.AccountLiquidationStatus(v.Status), Reason: v.Reason,
		StartedAt: v.StartedAt, CompletedAt: v.CompletedAt, Version: v.Version,
		CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes,
	}
}
func AccountLiquidationItemProto(v *models.TContractAccountLiquidationItem) *trade.ContractAccountLiquidationItem {
	return &trade.ContractAccountLiquidationItem{
		Id: v.Id, AccountLiquidationId: v.AccountLiquidationId, LiquidationNo: v.LiquidationNo,
		PositionId: v.PositionId, PositionVersion: v.PositionVersion, SymbolId: v.SymbolId,
		PositionSide: trade.PositionSide(v.PositionSide), TriggerQty: v.TriggerQty.String(),
		TriggerMarkPrice: v.TriggerMarkPrice.String(), TriggerSnapshotId: v.TriggerSnapshotId,
		PositionMargin: v.PositionMargin.String(), MaintenanceMargin: v.MaintenanceMargin.String(),
		RealizedPnl: v.RealizedPnl.String(), LiquidationFee: v.LiquidationFee.String(),
		DeficitAmount: v.DeficitAmount.String(), BankruptcyPrice: v.BankruptcyPrice.String(),
		AdlReliefAmount: v.AdlReliefAmount.String(), AdlQty: v.AdlQty.String(),
		Status: v.Status, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes,
	}
}
func SecondsPriceProto(v *models.TTradeSecondsPriceSnapshot) *trade.TradeSecondsPriceSnapshot {
	raw := ""
	if v.RawPayload.Valid {
		raw = v.RawPayload.String
	}
	return &trade.TradeSecondsPriceSnapshot{Id: v.Id, OrderId: v.OrderId, SnapshotType: trade.SecondsPriceSnapshotType(v.SnapshotType), Source: v.Source, Price: v.Price.String(), QuoteTime: v.QuoteTime, ReceivedAt: v.ReceivedAt, Algorithm: v.Algorithm, IsSelected: v.IsSelected, RawPayload: raw, CreateTimes: v.CreateTimes}
}
func ReservationProto(v *models.TTradeAssetReservation) *trade.TradeAssetReservation {
	return &trade.TradeAssetReservation{Id: v.Id, OrderId: v.OrderId, ReservationNo: v.ReservationNo, Asset: v.Asset, ReservedAmount: v.ReservedAmount.String(), ConsumedAmount: v.ConsumedAmount.String(), ReleasedAmount: v.ReleasedAmount.String(), Status: trade.AssetReservationStatus(v.Status), RetryCount: v.RetryCount, NextRetryAt: v.NextRetryAt, LastErrorMsg: v.LastErrorMsg, Version: v.Version, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}
func InstructionProto(v *models.TTradeSettlementInstruction) *trade.TradeSettlementInstruction {
	return &trade.TradeSettlementInstruction{Id: v.Id, InstructionNo: v.InstructionNo, BizType: v.BizType, BizId: v.BizId, BatchNo: v.BatchNo, FillId: v.FillId, OrderId: v.OrderId, PositionId: v.PositionId, ReservationNo: v.ReservationNo, UserId: v.UserId, Action: trade.SettlementInstructionAction(v.Action), Asset: v.Asset, Amount: v.Amount.String(), StepNo: v.StepNo, Status: trade.SettlementInstructionStatus(v.Status), RetryCount: v.RetryCount, NextRetryAt: v.NextRetryAt, LastErrorMsg: v.LastErrorMsg, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes, AssetFlowNo: v.AssetFlowNo, ReconciledAt: v.ReconciledAt}
}

func ContractReconciliationIssueProto(v *models.TContractReconciliationIssue) *trade.ContractReconciliationIssue {
	return &trade.ContractReconciliationIssue{
		Id:               v.Id,
		IssueKey:         v.IssueKey,
		CheckType:        v.CheckType,
		BizType:          v.BizType,
		BizNo:            v.BizNo,
		InstructionId:    v.InstructionId,
		ExpectedValue:    v.ExpectedValue,
		ActualValue:      v.ActualValue,
		Detail:           v.Detail,
		Status:           trade.ContractReconciliationIssueStatus(v.Status),
		OccurrenceCount:  v.OccurrenceCount,
		FirstSeenAt:      v.FirstSeenAt,
		LastSeenAt:       v.LastSeenAt,
		ResolvedAt:       v.ResolvedAt,
		OperatorId:       v.OperatorId,
		ResolutionReason: v.ResolutionReason,
		CreateTimes:      v.CreateTimes,
		UpdateTimes:      v.UpdateTimes,
	}
}
