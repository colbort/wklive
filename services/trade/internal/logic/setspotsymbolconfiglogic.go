package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	if errors.Is(err, models.ErrNotFound) || (err == nil && symbol.TenantId != in.TenantId) {
		return &trade.AdminCommonResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	cfg, err := l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, in.TenantId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.TTradeSymbolSpot{
			TenantId:    in.TenantId,
			SymbolId:    in.SymbolId,
			CreateTimes: now,
		}
	}
	cfg.MakerFeeRate = mustParseFloat(in.MakerFeeRate)
	cfg.TakerFeeRate = mustParseFloat(in.TakerFeeRate)
	cfg.BuyEnabled = conv.BoolInt64(in.BuyEnabled)
	cfg.SellEnabled = conv.BoolInt64(in.SellEnabled)
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
