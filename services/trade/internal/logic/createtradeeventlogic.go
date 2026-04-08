package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTradeEventLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTradeEventLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradeEventLogic {
	return &CreateTradeEventLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建交易事件
func (l *CreateTradeEventLogic) CreateTradeEvent(in *trade.CreateTradeEventReq) (*trade.InternalCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.InternalCommonResp{}, nil
}
