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
	if symbol.ProductType != int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE) || in.MarginMode == trade.MarginMode_MARGIN_MODE_UNKNOWN || in.Leverage <= 0 {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	now := utils.NowMillis()
	item, err := l.svcCtx.SymbolLeverageCfgModel.FindOneByTenantIdSymbolIdMarginModeLeverage(l.ctx, symbol.TenantId, in.SymbolId, int64(in.MarginMode), in.Leverage)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TTradeSymbolLeverageConfig{
			TenantId:    symbol.TenantId,
			SymbolId:    in.SymbolId,
			MarginMode:  int64(in.MarginMode),
			Leverage:    in.Leverage,
			Enabled:     int64(common.Enable_ENABLE_ENABLED),
			CreateTimes: now,
		}
	}

	if in.Enabled > 0 {
		item.Enabled = enableToModel(in.Enabled, item.Enabled)
	}
	if in.Sort > 0 {
		item.Sort = in.Sort
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	item.UpdateTimes = now

	if item.Id == 0 {
		if _, err = l.svcCtx.SymbolLeverageCfgModel.Insert(l.ctx, item); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.SymbolLeverageCfgModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	if in.IsDefault != common.YesNo_YES_NO_UNKNOWN {
		defaultItem, findErr := l.svcCtx.SymbolLeverageDefaultModel.FindOneByTenantIdSymbolIdMarginMode(l.ctx, symbol.TenantId, in.SymbolId, int64(in.MarginMode))
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if in.IsDefault == common.YesNo_YES_NO_YES {
			if item.Enabled != int64(common.Enable_ENABLE_ENABLED) {
				return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
			}
			if defaultItem == nil {
				defaultItem = &models.TTradeSymbolLeverageDefault{TenantId: symbol.TenantId, SymbolId: in.SymbolId, MarginMode: int64(in.MarginMode), CreateTimes: now}
			}
			defaultItem.Leverage, defaultItem.UpdateTimes = in.Leverage, now
			if defaultItem.Id == 0 {
				_, err = l.svcCtx.SymbolLeverageDefaultModel.Insert(l.ctx, defaultItem)
			} else {
				err = l.svcCtx.SymbolLeverageDefaultModel.Update(l.ctx, defaultItem)
			}
		} else if defaultItem != nil && defaultItem.Leverage == in.Leverage {
			err = l.svcCtx.SymbolLeverageDefaultModel.Delete(l.ctx, defaultItem.Id)
		}
		if err != nil {
			return nil, err
		}
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
