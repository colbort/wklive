package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserUpdateLogic {
	return &SysUserUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserUpdateLogic) SysUserUpdate(in *system.SysUserUpdateReq) (*system.RespBase, error) {
	if in.Id == 1 {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.SuperAdminCannotBeDeleted, i18n.Translate(i18n.SuperAdminCannotBeDeleted, l.ctx)),
		}, nil
	}
	one, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if in.Nickname != "" {
		one.Nickname = in.Nickname
	}
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		one.Enabled = commonStatusToModel(in.Enabled)
	}
	if in.TenantId != 0 {
		one.TenantId = in.TenantId
	}
	if in.UserType != system.UserType_USER_TYPE_UNKNOWN {
		one.UserType = int64(in.UserType)
	}
	if in.IsOwner != common.YesNo_YES_NO_UNKNOWN {
		one.IsOwner = yesNoToModel(in.IsOwner)
	}
	if in.Avatar != "" {
		one.Avatar = in.Avatar
	}

	err = l.svcCtx.UserModel.Update(l.ctx, one)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
