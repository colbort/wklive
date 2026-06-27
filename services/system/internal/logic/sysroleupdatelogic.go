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
			Base: helper.ErrResp(i18n.RoleNotFound, i18n.Translate(i18n.RoleNotFound, l.ctx)),
		}, nil
	}

	if one.Code == "super_admin" || one.Code == "tenant_super_admin" {
		return nil, i18n.StatusError(l.ctx, i18n.SuperAdminUpdateForbidden)
	}
	allowTenantUpdate, base, err := adminTenantWriteScope(l.ctx, one.TenantId, i18n.RoleNotFound)
	if err != nil {
		return nil, err
	}
	if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	if in.Name != "" {
		one.Name = in.Name
	}
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		one.Enabled = commonStatusToModel(in.Enabled)
	}
	if in.Remark != "" {
		one.Remark = in.Remark
	}
	if allowTenantUpdate && in.TenantId != 0 {
		one.TenantId = in.TenantId
	}
	if in.Code != "" {
		one.Code = in.Code
	}

	err = l.svcCtx.RoleModel.Update(l.ctx, one)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
