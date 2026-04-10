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
	item, err := findOrderByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.OrderId, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppGetOrderDetailResp{Base: helper.GetErrResp(404, "订单不存在")}, nil
		}
		return nil, err
	}
	if item.Uid != in.Uid || item.AccountId != in.AccountId {
		return &option.AppGetOrderDetailResp{Base: helper.GetErrResp(403, "无权查看该订单")}, nil
	}
	data, err := buildOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.AppGetOrderDetailResp{Base: helper.OkResp(), Data: data}, nil
}
