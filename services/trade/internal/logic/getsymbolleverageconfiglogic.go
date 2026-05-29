package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolLeverageConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolLeverageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolLeverageConfigLogic {
	return &GetSymbolLeverageConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对杠杆档位配置详情
func (l *GetSymbolLeverageConfigLogic) GetSymbolLeverageConfig(in *trade.GetSymbolLeverageConfigReq) (*trade.GetSymbolLeverageConfigResp, error) {
	var (
		item *models.TTradeSymbolLeverageConfig
		err  error
	)
	if in.Id > 0 {
		item, err = l.svcCtx.SymbolLeverageCfgModel.FindOne(l.ctx, in.Id)
	} else {
		item, err = l.svcCtx.SymbolLeverageCfgModel.FindOneByTenantIdSymbolIdMarketTypeMarginMode(l.ctx, in.TenantId, in.SymbolId, int64(in.MarketType), int64(in.MarginMode))
	}
	if errors.Is(err, models.ErrNotFound) || (err == nil && item.TenantId != in.TenantId) {
		return &trade.GetSymbolLeverageConfigResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}

	return &trade.GetSymbolLeverageConfigResp{
		Base: helper.OkResp(),
		Data: symbolLeverageConfigToProto(item),
	}, nil
}
