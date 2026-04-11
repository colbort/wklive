package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"sort"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"
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
func (l *GetProfileLogic) GetProfile(in *system.ProfileReq) (*system.ProfileResp, error) {
	// 1) user info
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	if u.Status != 1 {
		return nil, errors.New(i18n.Translate(i18n.UserDisabled, l.ctx))
	}

	// 2) roleIds
	roleIds, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserId(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	if len(roleIds) == 0 {
		// 没角色：只返回首页/空菜单
		return &system.ProfileResp{
			User: &system.ProfileUser{
				Id:       u.Id,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			},
			Menus:   []*system.SysMenuNode{},
			Perms:   []string{},
			RoleIds: []int64{},
		}, nil
	}

	// 3) menuIds
	menuIds, err := l.svcCtx.RoleMenuModel.FindMenuIdsByRoleIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return &system.ProfileResp{
			User: &system.ProfileUser{
				Id:       u.Id,
				Username: u.Username,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
			},
			Menus:   []*system.SysMenuNode{},
			Perms:   []string{},
			RoleIds: roleIds,
		}, nil
	}

	// 4) menus flat
	menus, err := l.svcCtx.MenuModel.FindByIds(l.ctx, menuIds, 1, 1)
	if err != nil {
		return nil, err
	}

	// 5) build tree + perms
	tree, perms := buildMenuTreeAndPerms(menus)

	return &system.ProfileResp{
		User: &system.ProfileUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
		},
		Menus:   tree,
		Perms:   perms,
		RoleIds: roleIds,
	}, nil
}

func buildMenuTreeAndPerms(rows []*models.SysMenu) ([]*system.SysMenuNode, []string) {
	nodes := make(map[int64]*system.SysMenuNode, len(rows))
	childrenMap := make(map[int64][]*system.SysMenuNode)
	permsSet := map[string]struct{}{}

	// 1. 创建节点；按钮只收集 perms，不放入 tree
	for _, r := range rows {
		// 收集按钮权限
		if r.MenuType == 3 {
			if r.Perms != "" {
				permsSet[r.Perms] = struct{}{}
			}
			continue
		}

		n := &system.SysMenuNode{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Name:      r.Name,
			MenuType:  r.MenuType,
			Path:      r.Path,
			Component: r.Component,
			Icon:      r.Icon,
			Sort:      r.Sort,
			Visible:   r.Visible,
			Status:    r.Status,
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
