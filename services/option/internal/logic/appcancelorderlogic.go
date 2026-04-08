package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppCancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCancelOrderLogic {
	return &AppCancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销期权委托订单
func (l *AppCancelOrderLogic) AppCancelOrder(in *option.AppCancelOrderReq) (*option.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppCommonResp{}, nil
}
