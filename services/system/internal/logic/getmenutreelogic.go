package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
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
	mm := make(map[int64]bool, len(menus))
	tenantId, _ := utils.GetTenantIdFromMd(l.ctx)
	filter := false
	if tenantId > 0 || in.TenantId > 0 {
		roleMenus, err := l.svcCtx.RoleMenuModel.ListByRoleId(l.ctx, 2)
		if err != nil {
			return nil, err
		}
		for _, v := range roleMenus {
			mm[v.MenuId] = true
		}
		filter = true
	}

	data := make([]*system.SysMenuItem, 0, len(menus))
	for _, m := range menus {
		if !mm[m.Id] && filter {
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
			Status:    commonStatusToProto(m.Status),
			Perms:     m.Perms,
		}
		data = append(data, item)
	}

	return &system.SysMenuTreeResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
