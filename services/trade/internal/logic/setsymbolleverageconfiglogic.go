package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSymbolLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSymbolLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSymbolLeverageConfigLogic {
	return &SetSymbolLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置交易对杠杆档位配置
func (l *SetSymbolLeverageConfigLogic) SetSymbolLeverageConfig(in *trade.SetSymbolLeverageConfigReq) (*trade.AdminCommonResp, error) {
	l.Errorf("set symbol leverage config started, tenantId=%d symbolId=%d marginMode=%d leverageValues=%v defaultLeverage=%d enabled=%d sort=%d",
		in.TenantId, in.SymbolId, in.MarginMode, in.LeverageValues, in.DefaultLeverage, in.Enabled, in.Sort)

	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		l.Infof("set symbol leverage config rejected: symbol not found, tenantId=%d symbolId=%d", in.TenantId, in.SymbolId)
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		l.Errorf("set symbol leverage config query symbol failed, tenantId=%d symbolId=%d err=%v", in.TenantId, in.SymbolId, err)
		return nil, err
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		l.Errorf("set symbol leverage config tenant scope check failed, requestTenantId=%d symbolTenantId=%d symbolId=%d err=%v", in.TenantId, symbol.TenantId, symbol.Id, err)
		return nil, err
	} else if base != nil {
		l.Infof("set symbol leverage config rejected: tenant out of scope, requestTenantId=%d symbolTenantId=%d symbolId=%d", in.TenantId, symbol.TenantId, symbol.Id)
		return &trade.AdminCommonResp{Base: base}, nil
	}
	if symbol.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		l.Errorf("set symbol leverage config rejected: product is not derivative, tenantId=%d symbolId=%d productType=%d", symbol.TenantId, symbol.Id, symbol.ProductType)
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	if len(in.LeverageValues) == 0 || in.MarginMode == trade.MarginMode_MARGIN_MODE_UNKNOWN || in.DefaultLeverage <= 0 || (in.Enabled != common.Enable_ENABLE_ENABLED && in.Enabled != common.Enable_ENABLE_DISABLED) || in.Sort < 0 {
		l.Errorf("set symbol leverage config rejected: invalid group parameters, tenantId=%d symbolId=%d valuesCount=%d marginMode=%d defaultLeverage=%d enabled=%d sort=%d",
			symbol.TenantId, symbol.Id, len(in.LeverageValues), in.MarginMode, in.DefaultLeverage, in.Enabled, in.Sort)
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	values := make([]int64, 0, len(in.LeverageValues))
	seen := make(map[int64]struct{}, len(in.LeverageValues))
	defaultFound := false
	for _, leverage := range in.LeverageValues {
		if leverage <= 0 {
			l.Errorf("set symbol leverage config rejected: non-positive leverage, tenantId=%d symbolId=%d leverage=%d", symbol.TenantId, symbol.Id, leverage)
			return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		if _, exists := seen[leverage]; exists {
			continue
		}
		seen[leverage] = struct{}{}
		values = append(values, leverage)
		defaultFound = defaultFound || leverage == in.DefaultLeverage
	}
	if !defaultFound {
		l.Errorf("set symbol leverage config rejected: default leverage is not selectable, tenantId=%d symbolId=%d defaultLeverage=%d leverageValues=%v",
			symbol.TenantId, symbol.Id, in.DefaultLeverage, values)
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	contractCfg, findErr := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, symbol.Id)
	if findErr != nil {
		if errors.Is(findErr, models.ErrNotFound) {
			l.Errorf("set symbol leverage config rejected: contract config not found, tenantId=%d symbolId=%d", symbol.TenantId, symbol.Id)
			return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		l.Errorf("set symbol leverage config query contract config failed, tenantId=%d symbolId=%d err=%v", symbol.TenantId, symbol.Id, findErr)
		return nil, findErr
	}
	if (in.MarginMode == trade.MarginMode_MARGIN_MODE_CROSS && contractCfg.SupportCross != int64(common.Enable_ENABLE_ENABLED)) ||
		(in.MarginMode == trade.MarginMode_MARGIN_MODE_ISOLATED && contractCfg.SupportIsolated != int64(common.Enable_ENABLE_ENABLED)) {
		l.Errorf("set symbol leverage config rejected: margin mode unsupported, tenantId=%d symbolId=%d marginMode=%d supportCross=%d supportIsolated=%d",
			symbol.TenantId, symbol.Id, in.MarginMode, contractCfg.SupportCross, contractCfg.SupportIsolated)
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if err := l.svcCtx.SymbolLeverageCfgModel.SyncGroup(l.ctx, symbol.TenantId, symbol.Id, int64(in.MarginMode), values, in.DefaultLeverage, int64(in.Enabled), in.Sort, in.Remark, utils.NowMillis()); err != nil {
		l.Errorf("set symbol leverage config sync group failed, tenantId=%d symbolId=%d marginMode=%d leverageValues=%v defaultLeverage=%d err=%v",
			symbol.TenantId, symbol.Id, in.MarginMode, values, in.DefaultLeverage, err)
		return nil, err
	}
	l.Errorf("set symbol leverage config succeeded, tenantId=%d symbolId=%d marginMode=%d leverageValues=%v defaultLeverage=%d enabled=%d sort=%d",
		symbol.TenantId, symbol.Id, in.MarginMode, values, in.DefaultLeverage, in.Enabled, in.Sort)
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
