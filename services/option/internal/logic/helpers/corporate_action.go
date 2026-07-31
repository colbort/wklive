package helpers

import (
	"errors"
	"strings"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

var ErrCorporateActionInexact = errors.New("corporate action conversion is not exactly representable")

type CorporateActionPositionConversion struct {
	SuccessorQuantity          decimal.Decimal
	SuccessorAvailableQuantity decimal.Decimal
	SuccessorOpenAvgPrice      decimal.Decimal
	SourceEffectiveMultiplier  decimal.Decimal
	TargetEffectiveMultiplier  decimal.Decimal
	CostBasisBefore            decimal.Decimal
	CostBasisAfter             decimal.Decimal
}

func ParsePositiveCorporateActionInteger(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, ErrCorporateActionInexact
	}
	result, err := decimal.NewFromString(value)
	if err != nil || !result.IsPositive() || !result.Equal(result.Truncate(0)) {
		return decimal.Zero, ErrCorporateActionInexact
	}
	if len(result.StringFixed(0)) > 32 {
		return decimal.Zero, ErrCorporateActionInexact
	}
	return result, nil
}

func OptionEffectiveMultiplier(contract *models.TOptionContract) decimal.Decimal {
	if contract != nil && contract.Multiplier.IsPositive() {
		return contract.Multiplier
	}
	if contract != nil && contract.ContractUnit.IsPositive() {
		return contract.ContractUnit
	}
	return decimal.NewFromInt(1)
}

func ExactCorporateActionQuantity(
	value, numerator, denominator decimal.Decimal,
) (decimal.Decimal, error) {
	if value.IsNegative() || !numerator.IsPositive() || !denominator.IsPositive() {
		return decimal.Zero, ErrCorporateActionInexact
	}
	result := value.Mul(numerator).Div(denominator)
	if !result.Equal(result.Truncate(16)) {
		return decimal.Zero, ErrCorporateActionInexact
	}
	return result, nil
}

func ConvertCorporateActionPosition(
	position *models.TOptionPosition,
	source, successor *models.TOptionContract,
	numerator, denominator decimal.Decimal,
) (*CorporateActionPositionConversion, error) {
	if position == nil || source == nil || successor == nil ||
		!position.PositionQty.IsPositive() || position.FrozenQty.IsPositive() {
		return nil, ErrCorporateActionInexact
	}
	successorQty, err := ExactCorporateActionQuantity(position.PositionQty, numerator, denominator)
	if err != nil || !successorQty.IsPositive() {
		return nil, ErrCorporateActionInexact
	}
	if successor.QtyStep.IsPositive() && !successorQty.Mod(successor.QtyStep).IsZero() {
		return nil, ErrCorporateActionInexact
	}
	successorAvailable, err := ExactCorporateActionQuantity(position.AvailableQty, numerator, denominator)
	if err != nil {
		return nil, ErrCorporateActionInexact
	}
	if successor.QtyStep.IsPositive() && !successorAvailable.Mod(successor.QtyStep).IsZero() {
		return nil, ErrCorporateActionInexact
	}
	sourceMultiplier := OptionEffectiveMultiplier(source)
	successorMultiplier := OptionEffectiveMultiplier(successor)
	costBefore := position.PositionQty.Mul(position.OpenAvgPrice).Mul(sourceMultiplier).Round(16)
	denominatorValue := successorQty.Mul(successorMultiplier)
	if !denominatorValue.IsPositive() {
		return nil, ErrCorporateActionInexact
	}
	successorAvg := costBefore.Div(denominatorValue)
	if !successorAvg.Equal(successorAvg.Truncate(16)) {
		return nil, ErrCorporateActionInexact
	}
	costAfter := successorQty.Mul(successorAvg).Mul(successorMultiplier).Round(16)
	if !costBefore.Equal(costAfter) {
		return nil, ErrCorporateActionInexact
	}
	return &CorporateActionPositionConversion{
		SuccessorQuantity: successorQty, SuccessorAvailableQuantity: successorAvailable,
		SuccessorOpenAvgPrice: successorAvg, SourceEffectiveMultiplier: sourceMultiplier,
		TargetEffectiveMultiplier: successorMultiplier, CostBasisBefore: costBefore, CostBasisAfter: costAfter,
	}, nil
}

func ToCorporateActionContractProto(item *models.TOptionCorporateActionContract) *option.OptionCorporateActionContract {
	if item == nil {
		return nil
	}
	return &option.OptionCorporateActionContract{
		Id: item.Id, TenantId: item.TenantId, ActionId: item.ActionId,
		SourceContractId: item.SourceContractId, SuccessorContractId: item.SuccessorContractId,
		ExecutionMode:       option.CorporateActionExecutionMode(item.ExecutionMode),
		QuantityNumerator:   item.QuantityNumerator.StringFixed(0),
		QuantityDenominator: item.QuantityDenominator.StringFixed(0),
		HaltId:              item.HaltId, Status: option.CorporateActionContractStatus(item.Status),
		PositionTotal: item.PositionTotal, PositionCompleted: item.PositionCompleted,
		PositionFailed: item.PositionFailed, LastPositionId: item.LastPositionId,
		RetryCount: item.RetryCount, LastErrorMsg: item.LastErrorMsg,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}

func ToCorporateActionProto(
	item *models.TOptionCorporateAction,
	contracts []*models.TOptionCorporateActionContract,
) *option.OptionCorporateAction {
	if item == nil {
		return nil
	}
	result := &option.OptionCorporateAction{
		Id: item.Id, TenantId: item.TenantId, EventNo: item.EventNo,
		ExternalEventRef: item.ExternalEventRef, Version: item.Version,
		UnderlyingSymbol: item.UnderlyingSymbol, ActionType: option.CorporateActionType(item.ActionType),
		Status: option.CorporateActionStatus(item.Status), AnnouncementTime: item.AnnouncementTime,
		ExTime: item.ExTime, RecordTime: item.RecordTime, EffectiveTime: item.EffectiveTime,
		PayTime: item.PayTime, EvidenceRef: item.EvidenceRef, EvidenceHash: item.EvidenceHash,
		Description: item.Description, CreatedBy: item.CreatedBy, ReviewedBy: item.ReviewedBy,
		ReviewReason: item.ReviewReason, ReviewedAt: item.ReviewedAt, CompletedAt: item.CompletedAt,
		LastErrorMsg: item.LastErrorMsg, CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
		Contracts: make([]*option.OptionCorporateActionContract, 0, len(contracts)),
	}
	for _, contract := range contracts {
		result.Contracts = append(result.Contracts, ToCorporateActionContractProto(contract))
	}
	return result
}

func ToCorporateActionPositionProto(item *models.TOptionCorporateActionPosition) *option.OptionCorporateActionPosition {
	if item == nil {
		return nil
	}
	return &option.OptionCorporateActionPosition{
		Id: item.Id, TenantId: item.TenantId, ActionId: item.ActionId,
		ActionContractId: item.ActionContractId, SourcePositionId: item.SourcePositionId,
		SuccessorPositionId: item.SuccessorPositionId, UserId: item.UserId, AccountId: item.AccountId,
		Side: int32(item.Side), SourceQuantity: item.SourceQuantity.String(),
		SuccessorQuantity:            item.SuccessorQuantity.String(),
		SourceAvailableQuantity:      item.SourceAvailableQuantity.String(),
		SuccessorAvailableQuantity:   item.SuccessorAvailableQuantity.String(),
		SourceOpenAvgPrice:           item.SourceOpenAvgPrice.String(),
		SuccessorOpenAvgPrice:        item.SuccessorOpenAvgPrice.String(),
		SourceEffectiveMultiplier:    item.SourceEffectiveMultiplier.String(),
		SuccessorEffectiveMultiplier: item.SuccessorEffectiveMultiplier.String(),
		CostBasisBefore:              item.CostBasisBefore.String(), CostBasisAfter: item.CostBasisAfter.String(),
		CashDifference: item.CashDifference.String(), Status: option.CorporateActionPositionStatus(item.Status),
		RetryCount: item.RetryCount, LastErrorMsg: item.LastErrorMsg, CompletedAt: item.CompletedAt,
		CreateTimes: item.CreateTimes, UpdateTimes: item.UpdateTimes,
	}
}
