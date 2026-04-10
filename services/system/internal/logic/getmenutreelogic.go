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

func (l *GetMenuTreeLogic) GetMenuTree(in *system.Empty) (*system.SysMenuTreeResp, error) {
	menus, err := l.svcCtx.MenuModel.ListAll(l.ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*system.SysMenuItem, 0, len(menus))
	for _, m := range menus {
		item := &system.SysMenuItem{
			Id:        m.Id,
			ParentId:  m.ParentId,
			Name:      m.Name,
			MenuType:  m.MenuType,
			Path:      m.Path,
			Component: m.Component,
			Icon:      m.Icon,
			Sort:      m.Sort,
			Visible:   m.Visible,
			Status:    m.Status,
			Perms:     m.Perms,
		}
		data = append(data, item)
	}

	return &system.SysMenuTreeResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
