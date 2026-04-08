package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListAdminLogic {
	return &GetOrderListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台订单列表
func (l *GetOrderListAdminLogic) GetOrderListAdmin(in *trade.GetOrderListAdminReq) (*trade.GetOrderListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetOrderListAdminResp{}, nil
}
