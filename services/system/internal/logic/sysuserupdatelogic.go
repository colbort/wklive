package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"
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
	var data models.SysUser
	_ = copier.Copy(&data, one)
	copyNonZero(&data, in)
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		data.Enabled = commonStatusToModel(in.Enabled)
	}

	err = l.svcCtx.UserModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
