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

type SysRoleUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleUpdateLogic) SysRoleUpdate(in *system.SysRoleUpdateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.RoleNotFound, i18n.Translate(i18n.RoleNotFound, l.ctx)),
		}, nil
	}

	if one.Code == "super_admin" || one.Code == "tenant_super_admin" {
		return nil, i18n.StatusError(l.ctx, i18n.SuperAdminUpdateForbidden)
	}

	var data models.SysRole
	_ = copier.Copy(&data, one)
	copyNonZero(&data, in)
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		data.Enabled = commonStatusToModel(in.Enabled)
	}

	err = l.svcCtx.RoleModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
