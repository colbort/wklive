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
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}
	if symbol.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	if len(in.LeverageValues) == 0 ||
		(in.MarginMode != trade.MarginMode_MARGIN_MODE_CROSS && in.MarginMode != trade.MarginMode_MARGIN_MODE_ISOLATED) ||
		in.DefaultLeverage <= 0 ||
		(in.Enabled != common.Enable_ENABLE_ENABLED && in.Enabled != common.Enable_ENABLE_DISABLED) ||
		in.Sort < 0 {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	values := make([]int64, 0, len(in.LeverageValues))
	seen := make(map[int64]struct{}, len(in.LeverageValues))
	defaultFound := false
	for _, leverage := range in.LeverageValues {
		if leverage <= 0 {
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
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	contractCfg, findErr := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, symbol.Id)
	if findErr != nil {
		if errors.Is(findErr, models.ErrNotFound) {
			return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		return nil, findErr
	}
	if !contractSupportsMarginMode(contractCfg, in.MarginMode) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, "该合约不支持所选保证金模式，不能配置对应杠杆")}, nil
	}
	if err := l.svcCtx.SymbolLeverageCfgModel.SyncGroup(l.ctx, symbol.TenantId, symbol.Id, int64(in.MarginMode), values, in.DefaultLeverage, int64(in.Enabled), in.Sort, in.Remark, utils.NowMillis()); err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
