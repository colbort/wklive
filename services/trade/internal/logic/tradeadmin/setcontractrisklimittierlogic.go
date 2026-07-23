package tradeadminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractRiskLimitTierLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractRiskLimitTierLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractRiskLimitTierLogic {
	return &SetContractRiskLimitTierLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存合约风险限额档位
func (l *SetContractRiskLimitTierLogic) SetContractRiskLimitTier(in *trade.SetContractRiskLimitTierReq) (*trade.CommonResp, error) {
	tenantID := adminTenantID(l.ctx, in.TenantId)
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if err != nil || symbol.TenantId != tenantID || symbol.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "derivative symbol not found")}, nil
	}
	floor, capValue := mustParseFloat(in.NotionalFloor), mustParseFloat(in.NotionalCap)
	initial, maintenance := mustParseFloat(in.InitialMarginRate), mustParseFloat(in.MaintenanceMarginRate)
	maintenanceAmount := mustParseFloat(in.MaintenanceAmount)
	if in.TierNo <= 0 || in.MaxLeverage <= 0 || floor.IsNegative() || (capValue.IsPositive() && !capValue.GreaterThan(floor)) || maintenance.IsNegative() || initial.LessThan(maintenance) || maintenanceAmount.IsNegative() {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "invalid risk tier")}, nil
	}
	var overlap int64
	query := "SELECT COUNT(1) FROM t_contract_risk_limit_tier WHERE tenant_id=? AND symbol_id=? AND id<>? AND tier_no<>? AND (notional_cap=0 OR notional_cap>?) AND (?=0 OR notional_floor<?)"
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &overlap, query, tenantID, in.SymbolId, in.Id, in.TierNo, floor, capValue, capValue); err != nil {
		return nil, err
	}
	if overlap > 0 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.ParamError, "risk tier notional range overlaps existing tier")}, nil
	}
	now, enabled := utils.NowMillis(), int64(in.Enabled)
	if enabled == 0 {
		enabled = 1
	}
	item := &models.TContractRiskLimitTier{TenantId: tenantID, SymbolId: in.SymbolId, TierNo: in.TierNo, NotionalFloor: floor, NotionalCap: capValue, MaxLeverage: in.MaxLeverage, InitialMarginRate: initial, MaintenanceMarginRate: maintenance, MaintenanceAmount: maintenanceAmount, Enabled: enabled, CreateTimes: now, UpdateTimes: now}
	if in.Id > 0 {
		existing, findErr := l.svcCtx.ContractRiskLimitTierModel.FindOne(l.ctx, in.Id)
		if errors.Is(findErr, models.ErrNotFound) || (findErr == nil && existing.TenantId != tenantID) {
			return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "risk tier not found")}, nil
		}
		if findErr != nil {
			return nil, findErr
		}
		item.Id, item.CreateTimes = existing.Id, existing.CreateTimes
		err = l.svcCtx.ContractRiskLimitTierModel.Update(l.ctx, item)
	} else {
		_, err = l.svcCtx.ContractRiskLimitTierModel.Insert(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
