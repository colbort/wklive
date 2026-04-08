package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAllOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllOrdersLogic {
	return &CancelAllOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销当前用户全部订单
func (l *CancelAllOrdersLogic) CancelAllOrders(in *trade.CancelAllOrdersReq) (*trade.CancelAllOrdersResp, error) {
	// todo: add your logic here and delete this line

	return &trade.CancelAllOrdersResp{}, nil
}
