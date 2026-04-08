package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetOrderLogic {
	return &AdminGetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个委托订单详情
func (l *AdminGetOrderLogic) AdminGetOrder(in *option.GetOrderReq) (*option.GetOrderResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetOrderResp{}, nil
}
