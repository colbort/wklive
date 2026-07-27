package applogic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个委托订单详情
func (l *GetOrderDetailLogic) GetOrderDetail(in *option.GetOrderDetailReq) (*option.GetOrderDetailResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := helpers.FindOrderByNoOrID(l.ctx, l.svcCtx, tenantId, in.OrderId, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetOrderDetailResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.UserId != userId || item.AccountId != in.AccountId {
		return &option.GetOrderDetailResp{Base: helper.ErrResp(i18n.NoPermissionViewOrder, i18n.Translate(i18n.NoPermissionViewOrder, l.ctx))}, nil
	}
	data, err := helpers.BuildOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetOrderDetailResp{Base: helper.OkResp(), Data: data}, nil
}
