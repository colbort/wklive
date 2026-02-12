package logic

import (
	"context"
	"strings"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleGrantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleGrantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantDetailLogic {
	return &SysRoleGrantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取角色授权详情
func (l *SysRoleGrantDetailLogic) SysRoleGrantDetail(in *system.SysRoleGrantDetailReq) (*system.SysRoleGrantDetailResp, error) {
	roleMenus, err := l.svcCtx.RoleMenuModel.ListByRoleId(l.ctx, in.RoleId)
	if err != nil {
		return nil, err
	}

	menuIds := make([]int64, 0, len(roleMenus))
	for _, rm := range roleMenus {
		menuIds = append(menuIds, rm.MenuId)
	}

	permKeys, err := l.permKeysFromMenuIds(menuIds)
	if err != nil {
		return nil, err
	}

	return &system.SysRoleGrantDetailResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "success",
		},
		RoleId:   in.RoleId,
		MenuIds:  menuIds,
		PermKeys: permKeys,
	}, nil
}

func (l *SysRoleGrantDetailLogic) permKeysFromMenuIds(menuIds []int64) ([]string, error) {
	if len(menuIds) == 0 {
		return []string{}, nil
	}

	permSet := make(map[string]struct{}, 64)

	for _, id := range menuIds {
		menu, err := l.svcCtx.MenuModel.FindOne(l.ctx, id)
		if err != nil {
			return nil, err
		}
		if menu == nil {
			continue
		}
		// 只取按钮
		if int32(menu.MenuType) != 3 {
			continue
		}
		key := strings.TrimSpace(menu.Perms)
		if key == "" {
			continue
		}
		permSet[key] = struct{}{}
	}

	out := make([]string, 0, len(permSet))
	for k := range permSet {
		out = append(out, k)
	}
	return out, nil
}
