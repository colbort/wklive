package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := findOrderByNoOrID(l.ctx, l.svcCtx, tenantId, in.OrderId, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppGetOrderDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.UserId != userId || item.AccountId != in.AccountId {
		return &option.AppGetOrderDetailResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionViewOrder, l.ctx))}, nil
	}
	data, err := buildOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.AppGetOrderDetailResp{Base: helper.OkResp(), Data: data}, nil
}
