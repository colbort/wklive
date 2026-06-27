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

type SetSpotSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSpotSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSpotSymbolConfigLogic {
	return &SetSpotSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置现货交易对配置
func (l *SetSpotSymbolConfigLogic) SetSpotSymbolConfig(in *trade.SetSpotSymbolConfigReq) (*trade.AdminCommonResp, error) {
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
	now := utils.NowMillis()
	cfg, err := l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, symbol.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.TTradeSymbolSpot{
			TenantId:    symbol.TenantId,
			SymbolId:    in.SymbolId,
			BuyEnabled:  int64(common.Enable_ENABLE_ENABLED),
			SellEnabled: int64(common.Enable_ENABLE_ENABLED),
			CreateTimes: now,
		}
	}
	cfg.MakerFeeRate = mustParseFloat(in.MakerFeeRate)
	cfg.TakerFeeRate = mustParseFloat(in.TakerFeeRate)
	cfg.BuyEnabled = enableToModel(in.BuyEnabled, cfg.BuyEnabled)
	cfg.SellEnabled = enableToModel(in.SellEnabled, cfg.SellEnabled)
	cfg.UpdateTimes = now
	if cfg.Id == 0 {
		if _, err = l.svcCtx.TradeSymbolSpotModel.Insert(l.ctx, cfg); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.TradeSymbolSpotModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}

	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
