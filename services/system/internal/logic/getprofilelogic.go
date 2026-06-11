package logic

import (
	"context"
	"sort"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取当前用户信息
func (l *GetProfileLogic) GetProfile(in *system.Empty) (*system.ProfileResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	// 1) user info
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	if u.Enabled != 1 {
		return nil, i18n.StatusError(l.ctx, i18n.UserDisabled)
	}

	// 2) roleIds
	roleIds, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserId(l.ctx, userId)
	if err != nil {
		return nil, err
	}
	if len(roleIds) == 0 {
		// 没角色：只返回首页/空菜单
		return profileResp(u, []*system.SysMenuNode{}, []string{}, []int64{}), nil
	}

	// 3) menuIds
	menuIds, err := l.svcCtx.RoleMenuModel.FindMenuIdsByRoleIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return profileResp(u, []*system.SysMenuNode{}, []string{}, roleIds), nil
	}

	// 4) menus flat
	menus, err := l.svcCtx.MenuModel.FindByIds(l.ctx, menuIds, 1, 1)
	if err != nil {
		return nil, err
	}

	// 5) build tree + perms
	tree, perms := buildMenuTreeAndPerms(menus)

	return profileResp(u, tree, perms, roleIds), nil
}

func profileResp(u *models.SysUser, menus []*system.SysMenuNode, perms []string, roleIds []int64) *system.ProfileResp {
	return &system.ProfileResp{
		Base: helper.OkResp(),
		Data: &system.ProfileData{
			User: &system.ProfileUser{
				Id:       u.Id,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
				TenantId: u.TenantId,
				UserType: system.UserType(u.UserType),
				IsOwner:  u.IsOwner,
			},
			Menus:   menus,
			Perms:   perms,
			RoleIds: roleIds,
		},
	}
}

func buildMenuTreeAndPerms(rows []*models.SysMenu) ([]*system.SysMenuNode, []string) {
	nodes := make(map[int64]*system.SysMenuNode, len(rows))
	childrenMap := make(map[int64][]*system.SysMenuNode)
	permsSet := map[string]struct{}{}

	// 1. 创建节点；按钮只收集 perms，不放入 tree
	for _, r := range rows {
		// 收集按钮权限
		if r.MenuType == int64(system.MenuType_MENU_TYPE_BUTTON) {
			if r.Perms != "" {
				permsSet[r.Perms] = struct{}{}
			}
			continue
		}

		n := &system.SysMenuNode{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Name:      r.Name,
			MenuType:  menuTypeToProto(r.MenuType),
			Path:      r.Path,
			Component: r.Component,
			Icon:      r.Icon,
			Sort:      r.Sort,
			Visible:   visibleStatusToProto(r.Visible),
			Enabled:   commonStatusToProto(r.Enabled),
			Perms:     r.Perms,
			Children:  nil,
		}
		nodes[r.Id] = n
	}

	// 2. 建立父子关系
	for _, n := range nodes {
		childrenMap[n.ParentId] = append(childrenMap[n.ParentId], n)
	}

	// 3. 排序
	sortChildren := func(list []*system.SysMenuNode) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Sort == list[j].Sort {
				return list[i].Id < list[j].Id
			}
			return list[i].Sort < list[j].Sort
		})
	}

	// 4. 递归组装树
	var build func(pid int64) []*system.SysMenuNode
	build = func(pid int64) []*system.SysMenuNode {
		list := childrenMap[pid]
		if len(list) == 0 {
			return nil
		}

		sortChildren(list)

		out := make([]*system.SysMenuNode, 0, len(list))
		for _, item := range list {
			item.Children = build(item.Id)
			out = append(out, item)
		}
		return out
	}

	tree := build(0)

	// 5. perms 转 slice
	perms := make([]string, 0, len(permsSet))
	for p := range permsSet {
		perms = append(perms, p)
	}
	sort.Strings(perms)

	return tree, perms
}
