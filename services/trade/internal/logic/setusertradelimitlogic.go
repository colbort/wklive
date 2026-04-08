package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserTradeLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserTradeLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTradeLimitLogic {
	return &SetUserTradeLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户交易限制
func (l *SetUserTradeLimitLogic) SetUserTradeLimit(in *trade.SetUserTradeLimitReq) (*trade.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.AdminCommonResp{}, nil
}
