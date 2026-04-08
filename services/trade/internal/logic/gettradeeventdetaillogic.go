package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradeEventDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradeEventDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventDetailLogic {
	return &GetTradeEventDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易事件详情
func (l *GetTradeEventDetailLogic) GetTradeEventDetail(in *trade.GetTradeEventDetailReq) (*trade.GetTradeEventDetailResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetTradeEventDetailResp{}, nil
}
