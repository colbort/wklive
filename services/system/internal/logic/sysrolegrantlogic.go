package logic

import (
	"context"
	"errors"
	"fmt"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	g "github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SysRoleGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantLogic {
	return &SysRoleGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleGrantLogic) SysRoleGrant(in *system.SysRoleGrantReq) (*system.RespBase, error) {
	menuIds := make([]int64, 0, len(in.MenuIds))
	seen := make(map[int64]struct{}, len(in.MenuIds))
	for _, id := range in.MenuIds {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		menuIds = append(menuIds, id)
	}

	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.RoleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}

	if len(menuIds) > 0 {
		existIds, err := l.svcCtx.MenuModel.FindIdsByIds(l.ctx, menuIds)
		if err != nil {
			return nil, err
		}
		if len(existIds) != len(menuIds) {
			existSet := make(map[int64]struct{}, len(existIds))
			for _, id := range existIds {
				existSet[id] = struct{}{}
			}
			missing := make([]int64, 0)
			for _, id := range menuIds {
				if _, ok := existSet[id]; !ok {
					missing = append(missing, id)
				}
			}
			return nil, fmt.Errorf("菜单不存在: %v", missing)
		}
	}

	err = l.svcCtx.RoleMenuModel.TransactCtx(l.ctx, func(ctx context.Context, session g.Session) error {
		if err := l.svcCtx.RoleMenuModel.DeleteByRoleId(ctx, in.RoleId); err != nil {
			return err
		}
		if len(menuIds) == 0 {
			return nil
		}
		rows := make([]*models.SysRoleMenu, 0, len(menuIds))
		for _, mid := range menuIds {
			rows = append(rows, &models.SysRoleMenu{
				RoleId: in.RoleId,
				MenuId: mid,
			})
		}
		return l.svcCtx.RoleMenuModel.InsertBatch(ctx, rows)
	})
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "授权成功",
	}, nil
}
