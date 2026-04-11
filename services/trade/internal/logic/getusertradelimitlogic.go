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

type GetUserTradeLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTradeLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTradeLimitLogic {
	return &GetUserTradeLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户交易限制
func (l *GetUserTradeLimitLogic) GetUserTradeLimit(in *trade.GetUserTradeLimitReq) (*trade.GetUserTradeLimitResp, error) {
	item, err := l.svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdMarketType(l.ctx, in.TenantId, in.UserId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserTradeLimitResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = riskUserTradeLimitToProto(item)
	}
	return resp, nil
}
