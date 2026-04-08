package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailAdminLogic {
	return &GetOrderDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取订单详情
func (l *GetOrderDetailAdminLogic) GetOrderDetailAdmin(in *trade.GetOrderDetailAdminReq) (*trade.GetOrderDetailAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetOrderDetailAdminResp{}, nil
}
