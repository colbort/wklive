package adminlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetUserTradingControlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetUserTradingControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetUserTradingControlLogic {
	return &AdminGetUserTradingControlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户 kill switch
func (l *AdminGetUserTradingControlLogic) AdminGetUserTradingControl(in *option.AdminGetUserTradingControlReq) (*option.AdminGetUserTradingControlResp, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.AdminGetUserTradingControlResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if in.TenantId <= 0 || in.UserId <= 0 {
		return &option.AdminGetUserTradingControlResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	item, err := l.svcCtx.OptionUserTradingControlModel.FindOneByTenantIdUserId(
		l.ctx, in.TenantId, in.UserId,
	)
	if errors.Is(err, models.ErrNotFound) {
		return &option.AdminGetUserTradingControlResp{
			Base: helper.OkResp(),
			Data: &option.OptionUserTradingControl{
				TenantId: in.TenantId, UserId: in.UserId, KillSwitch: common.YesNo_YES_NO_NO,
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &option.AdminGetUserTradingControlResp{
		Base: helper.OkResp(), Data: helpers.ToUserTradingControlProto(item),
	}, nil
}
