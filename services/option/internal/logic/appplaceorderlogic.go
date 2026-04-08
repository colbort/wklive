package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppPlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPlaceOrderLogic {
	return &AppPlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交期权下单请求
func (l *AppPlaceOrderLogic) AppPlaceOrder(in *option.AppPlaceOrderReq) (*option.AppPlaceOrderResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppPlaceOrderResp{}, nil
}
