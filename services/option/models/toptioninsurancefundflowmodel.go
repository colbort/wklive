package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/proto/option"
)

var _ TOptionInsuranceFundFlowModel = (*customTOptionInsuranceFundFlowModel)(nil)

// Raw insurance-fund flow amounts are positive business magnitudes. The flow
// type is the sole source of direction, which also makes legacy signed rows
// readable without rewriting their immutable economic evidence.
const optionInsuranceFundSignedAmountSQL = "CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END"

type (
	// TOptionInsuranceFundFlowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionInsuranceFundFlowModel.
	TOptionInsuranceFundFlowModel interface {
		tOptionInsuranceFundFlowModel
	}

	customTOptionInsuranceFundFlowModel struct {
		*defaultTOptionInsuranceFundFlowModel
	}
)

// NewTOptionInsuranceFundFlowModel returns a model for the database table.
func NewTOptionInsuranceFundFlowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionInsuranceFundFlowModel {
	return &customTOptionInsuranceFundFlowModel{
		defaultTOptionInsuranceFundFlowModel: newTOptionInsuranceFundFlowModel(conn, c, opts...),
	}
}

func insuranceFundSignedAmount(flowType int64, amount decimal.Decimal) (decimal.Decimal, error) {
	if !amount.IsPositive() {
		return decimal.Zero, errors.New("option insurance fund flow amount must be a positive magnitude")
	}
	switch option.InsuranceFundFlowType(flowType) {
	case option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_LIQUIDATION_FEE,
		option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_MANUAL_DEPOSIT:
		return amount, nil
	case option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER,
		option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_MANUAL_WITHDRAW:
		return amount.Neg(), nil
	default:
		return decimal.Zero, errors.New("invalid option insurance fund flow type")
	}
}

func (m *customTOptionInsuranceFundFlowModel) Insert(
	ctx context.Context, data *TOptionInsuranceFundFlow,
) (sql.Result, error) {
	if data == nil || data.TenantId <= 0 || strings.TrimSpace(data.FlowNo) == "" ||
		strings.TrimSpace(data.Coin) == "" || strings.TrimSpace(data.AssetFlowNo) == "" ||
		data.CreateTimes <= 0 {
		return nil, errors.New("invalid option insurance fund flow identity")
	}
	if _, err := insuranceFundSignedAmount(data.FlowType, data.Amount); err != nil {
		return nil, err
	}
	switch option.InsuranceFundFlowType(data.FlowType) {
	case option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_LIQUIDATION_FEE,
		option.InsuranceFundFlowType_INSURANCE_FUND_FLOW_TYPE_DEFICIT_COVER:
		if data.ContractId <= 0 || data.LiquidationId <= 0 {
			return nil, errors.New("liquidation insurance fund flow requires contract and liquidation")
		}
	}
	return m.defaultTOptionInsuranceFundFlowModel.Insert(ctx, data)
}
