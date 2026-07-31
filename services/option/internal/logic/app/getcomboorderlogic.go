package applogic

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

type GetComboOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComboOrderLogic {
	return &GetComboOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取组合订单及不可变腿
func (l *GetComboOrderLogic) GetComboOrder(in *option.GetComboOrderReq) (*option.GetComboOrderResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := findComboOrderByNoOrID(
		l.ctx, l.svcCtx, tenantID, in.ComboOrderId, in.ComboNo,
	)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetComboOrderResp{
				Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}
	if item.UserId != userID || item.AccountId != in.AccountId {
		return &option.GetComboOrderResp{
			Base: helper.ErrResp(
				i18n.NoPermissionOperateOrder,
				i18n.Translate(i18n.NoPermissionOperateOrder, l.ctx),
			),
		}, nil
	}
	detail, err := buildComboOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}
	return &option.GetComboOrderResp{Base: helper.OkResp(), Data: detail}, nil
}
