package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

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

func (l *GetUserTradeLimitLogic) GetUserTradeLimit(in *trade.GetUserTradeLimitReq) (*trade.GetUserTradeLimitResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetUserTradeLimitResp{}, nil
}
