package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuTreeLogic) GetMenuTree(in *system.SysMenuTreeReq) (*system.SysMenuTreeResp, error) {
	menus, err := l.svcCtx.MenuModel.ListAll(l.ctx)
	if err != nil {
		return nil, err
	}

	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.RoleId)
	if err != nil {
		return nil, err
	}
	roleId := int64(1)
	if role.TenantId > 0 || role.Code == "tenant_super_admin" || role.Code == "tenant_owner" {
		roleId = 2
	}
	mm := make(map[int64]bool, len(menus))
	roleMenus, err := l.svcCtx.RoleMenuModel.ListByRoleId(l.ctx, roleId)
	if err != nil {
		return nil, err
	}
	for _, v := range roleMenus {
		mm[v.MenuId] = true
	}

	data := make([]*system.SysMenuItem, 0, len(menus))
	for _, m := range menus {
		if !mm[m.Id] {
			continue
		}
		item := &system.SysMenuItem{
			Id:        m.Id,
			ParentId:  m.ParentId,
			Name:      m.Name,
			MenuType:  menuTypeToProto(m.MenuType),
			Path:      m.Path,
			Component: m.Component,
			Icon:      m.Icon,
			Sort:      m.Sort,
			Visible:   visibleStatusToProto(m.Visible),
			Enabled:   commonStatusToProto(m.Enabled),
			Perms:     m.Perms,
		}
		data = append(data, item)
	}

	return &system.SysMenuTreeResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
