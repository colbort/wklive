package logic

import (
	"context"
	"errors"
	"sort"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/errorx"
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
func (l *GetProfileLogic) GetProfile(in *system.ProfileReq) (*system.ProfileResp, error) {
	// 1) user info
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Uid)
	if err != nil {
		return nil, err
	}
	if u.Status != 1 {
		return nil, errorx.Wrap(errors.New("用户已禁用"), "用户已禁用")
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
			Menus: []*system.SysMenuNode{},
			Perms: []string{},
		}, nil
	}

	// 3) menuIds
	menuIds, err := l.svcCtx.RoleMenuModel.FindMenuIdsByRoleIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(menuIds) == 0 {
		return &system.ProfileResp{
			User:  &system.ProfileUser{Id: u.Id, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar},
			Menus: []*system.SysMenuNode{},
			Perms: []string{},
		}, nil
	}

	// 4) menus flat
	menus, err := l.svcCtx.MenuModel.FindByIds(l.ctx, menuIds)
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
		Menus: tree,
		Perms: perms,
	}, nil
}

func buildMenuTreeAndPerms(rows []*models.SysMenu) ([]*system.SysMenuNode, []string) {
	nodes := make(map[int64]*system.SysMenuNode, len(rows))
	childrenMap := make(map[int64][]*system.SysMenuNode)
	permsSet := map[string]struct{}{}

	// create nodes
	for _, r := range rows {
		n := &system.SysMenuNode{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Name:      r.Name,
			MenuType:  int32(r.MenuType),
			Path:      r.Path,
			Component: r.Component,
			Icon:      r.Icon,
			Sort:      int32(r.Sort),
			Visible:   int32(r.Visible),
			Status:    int32(r.Status),
			Perms:     r.Perms,
			Children:  []*system.SysMenuNode{},
		}
		nodes[r.Id] = n

		// perms from buttons
		if r.MenuType == 3 && r.Perms != "" {
			permsSet[r.Perms] = struct{}{}
		}
	}

	// link children
	for _, n := range nodes {
		childrenMap[n.ParentId] = append(childrenMap[n.ParentId], n)
	}

	// sort helper
	sortChildren := func(list []*system.SysMenuNode) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Sort == list[j].Sort {
				return list[i].Id < list[j].Id
			}
			return list[i].Sort < list[j].Sort
		})
	}

	// attach recursively
	var build func(pid int64) []*system.SysMenuNode
	build = func(pid int64) []*system.SysMenuNode {
		list := childrenMap[pid]
		sortChildren(list)
		out := make([]*system.SysMenuNode, 0, len(list))
		for _, item := range list {
			// 子节点
			child := build(item.Id)
			if len(child) > 0 {
				item.Children = child
			}
			out = append(out, item)
		}
		return out
	}

	tree := build(0)

	// perms to slice
	perms := make([]string, 0, len(permsSet))
	for p := range permsSet {
		perms = append(perms, p)
	}
	sort.Strings(perms)

	return tree, perms
}
