// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/rpc/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuTreeLogic {
	return &SysMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysMenuTreeLogic) SysMenuTree() (resp *types.SysMenuTreeResp, err error) {
	reuslt, err := l.svcCtx.SystemCli.GetMenuTree(l.ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysMenuItem, 0)
	for _, item := range reuslt.List {
		data = append(data, types.SysMenuItem{
			Id:        item.Id,
			ParentId:  item.ParentId,
			Name:      item.Name,
			MenuType:  item.MenuType,
			Icon:      item.Icon,
			Path:      item.Path,
			Component: item.Component,
			Sort:      item.Sort,
			Visible:   item.Visible,
			Status:    item.Status,
			Perms:     item.Perms,
		})
	}
	resp = &types.SysMenuTreeResp{
		RespBase: types.RespBase{
			Code: reuslt.Code,
			Msg:  reuslt.Msg,
		},
		Data: data,
	}
	return resp, nil
}
