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

type GetUserSymbolLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSymbolLimitLogic {
	return &GetUserSymbolLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户交易对限制
func (l *GetUserSymbolLimitLogic) GetUserSymbolLimit(in *trade.GetUserSymbolLimitReq) (*trade.GetUserSymbolLimitResp, error) {
	item, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolIdMarketType(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	resp := &trade.GetUserSymbolLimitResp{Base: helper.OkResp()}
	if item != nil {
		resp.Data = riskUserSymbolLimitToProto(item)
	}
	return resp, nil
}
