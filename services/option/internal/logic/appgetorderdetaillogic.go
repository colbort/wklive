package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppGetOrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetOrderDetailLogic {
	return &AppGetOrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个委托订单详情
func (l *AppGetOrderDetailLogic) AppGetOrderDetail(in *option.AppGetOrderDetailReq) (*option.AppGetOrderDetailResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppGetOrderDetailResp{}, nil
}
