package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

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
	// todo: add your logic here and delete this line

	return &trade.GetUserTradeConfigResp{}, nil
}
