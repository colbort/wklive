package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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
	item, err := findOrderByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetOrderResp{Base: helper.GetErrResp(404, "订单不存在")}, nil
		}
		return nil, err
	}
	data, err := buildOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetOrderResp{Base: helper.OkResp(), Data: data}, nil
}
