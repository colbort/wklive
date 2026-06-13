package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDetailLogic {
	return &SysUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserDetailLogic) SysUserDetail(in *system.SysUserDetailReq) (*system.SysUserDetailResp, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
		}
		return nil, err
	}
	if user == nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}

	roleIds, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserId(l.ctx, user.Id)
	if err != nil {
		return nil, err
	}

	return &system.SysUserDetailResp{
		Base: helper.OkResp(),
		Data: &system.SysUserItem{
			Id:               user.Id,
			Username:         user.Username,
			Nickname:         user.Nickname,
			Enabled:          commonStatusToProto(user.Enabled),
			RoleIds:          roleIds,
			CreateTimes:      user.CreateTimes,
			Google2FaEnabled: commonStatusToProto(user.GoogleEnabled),
			TenantId:         user.TenantId,
			UserType:         system.UserType(user.UserType),
			IsOwner:          common.YesNo(user.IsOwner),
			Avatar:           user.Avatar,
			PermsVer:         user.PermsVer,
			LastLoginIp:      user.LastLoginIp,
			LastLoginAt:      user.LastLoginAt,
			CreateBy:         user.CreateBy,
			UpdateTimes:      user.UpdateTimes,
		},
	}, nil
}
