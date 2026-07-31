package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminComboOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAdminComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminComboOrderLogic {
	return &GetAdminComboOrderLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

// 查询组合父单、腿、影子单、成交和资产指令
func (l *GetAdminComboOrderLogic) GetAdminComboOrder(
	in *option.GetAdminComboOrderReq,
) (*option.GetAdminComboOrderResp, error) {
	comboNo := strings.TrimSpace(in.ComboNo)
	if in.TenantId <= 0 || (in.Id <= 0 && comboNo == "") {
		return &option.GetAdminComboOrderResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.GetAdminComboOrderResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	item, err := findAdminComboOrder(l.ctx, l.svcCtx, in.TenantId, in.Id, comboNo)
	if errors.Is(err, models.ErrNotFound) {
		return &option.GetAdminComboOrderResp{
			Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	detail, err := buildAdminComboDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}
	return &option.GetAdminComboOrderResp{Base: helper.OkResp(), Data: detail}, nil
}
