package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleDeleteLogic {
	return &SysRoleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleDeleteLogic) SysRoleDelete(in *system.SysRoleDeleteReq) (*system.RespBase, error) {
	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, i18n.StatusError(l.ctx, i18n.RoleNotFound)
	}
	if role.Code == "super_admin" || role.Code == "tenant_super_admin" {
		return nil, i18n.StatusError(l.ctx, i18n.SuperAdminDeleteForbidden)
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, role.TenantId, i18n.RoleNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}
	err = l.svcCtx.RoleModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
