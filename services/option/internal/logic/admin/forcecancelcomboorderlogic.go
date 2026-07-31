package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForceCancelComboOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewForceCancelComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForceCancelComboOrderLogic {
	return &ForceCancelComboOrderLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

// 强制撤销一个组合父单；禁止单腿撤销
func (l *ForceCancelComboOrderLogic) ForceCancelComboOrder(
	in *option.ForceCancelComboOrderReq,
) (*option.CommonResp, error) {
	comboNo := strings.TrimSpace(in.ComboNo)
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || (in.Id <= 0 && comboNo == "") ||
		reason == "" || len(reason) > 200 {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	item, err := findAdminComboOrder(l.ctx, l.svcCtx, in.TenantId, in.Id, comboNo)
	if errors.Is(err, models.ErrNotFound) {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if err = applogic.CancelComboOrderByControl(
		l.ctx, l.svcCtx, item.Id, "ADMIN_FORCE_CANCEL:"+reason,
	); err != nil {
		return nil, err
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}
