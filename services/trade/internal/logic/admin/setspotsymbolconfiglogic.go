package adminlogic

import (
	"context"
	"errors"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
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
func (l *SetSpotSymbolConfigLogic) SetSpotSymbolConfig(in *trade.SetSpotSymbolConfigReq) (*trade.CommonResp, error) {
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
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
	cfg.MakerFeeRate = helpers.MustParseFloat(in.MakerFeeRate)
	cfg.TakerFeeRate = helpers.MustParseFloat(in.TakerFeeRate)
	cfg.BuyEnabled = helpers.EnableToModel(in.BuyEnabled, cfg.BuyEnabled)
	cfg.SellEnabled = helpers.EnableToModel(in.SellEnabled, cfg.SellEnabled)
	cfg.UpdateTimes = now
	if cfg.Id == 0 {
		if _, err = l.svcCtx.TradeSymbolSpotModel.Insert(l.ctx, cfg); err != nil {
			return nil, err
		}
	} else if err = l.svcCtx.TradeSymbolSpotModel.Update(l.ctx, cfg); err != nil {
		return nil, err
	}

	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
