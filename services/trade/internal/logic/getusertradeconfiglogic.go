package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTradeConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTradeConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradeConfigLogic {
	return &GetUserTradeConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户交易配置
func (l *GetUserTradeConfigLogic) GetUserTradeConfig(in *trade.GetUserTradeConfigReq) (*trade.GetUserTradeConfigResp, error) {
	item, err := l.svcCtx.TradeUserConfigModel.FindOneByTenantIdUserIdMarketTypeSymbolId(l.ctx, in.TenantId, in.UserId, int64(in.MarketType), in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserTradeConfigResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = userConfigToProto(item)
	}
	return resp, nil
}
