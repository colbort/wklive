package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradeEventListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradeEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventListLogic {
	return &GetTradeEventListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTradeEventListLogic) GetTradeEventList(in *trade.GetTradeEventListReq) (*trade.GetTradeEventListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetTradeEventListResp{}, nil
}
