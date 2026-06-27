package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetUserGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetUserGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserGoogle2FALogic {
	return &ResetUserGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置用户谷歌2FA
func (l *ResetUserGoogle2FALogic) ResetUserGoogle2FA(in *user.ResetUserGoogle2FAReq) (*user.AdminCommonResp, error) {
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, tuser.TenantId, i18n.NoPermissionOperateThisUser); err != nil {
		return nil, err
	} else if base != nil {
		return &user.AdminCommonResp{
			Base: base,
		}, nil
	}

	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AdminCommonResp{
			Base: helper.ErrResp(i18n.UserSecurityInfoNotFound, i18n.Translate(i18n.UserSecurityInfoNotFound, l.ctx)),
		}, nil
	}

	// 禁用 Google 2FA
	userSecurity.GoogleEnabled = int64(common.Enable_ENABLE_DISABLED)
	userSecurity.UpdateTimes = utils.NowMillis()
	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员重置用户 %d 的 Google2FA", in.UserId)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
