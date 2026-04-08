package logic

import (
	"context"
	"time"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuCreateLogic {
	return &SysMenuCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 菜单
func (l *SysMenuCreateLogic) SysMenuCreate(in *system.SysMenuCreateReq) (*system.RespBase, error) {
	var menu *models.SysMenu
	var err error
	switch in.MenuType {
	case 1:
		menu, err = l.svcCtx.MenuModel.FindOneByName(l.ctx, in.Name)
	case 2:
		menu, err = l.svcCtx.MenuModel.FindOneByPath(l.ctx, in.Path)
	case 3:
		menu, err = l.svcCtx.MenuModel.FindOneByPerms(l.ctx, in.Perms)
	default:
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "Invalid menu type",
			},
		}, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}
	if menu != nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "Menu already exists",
			},
		}, nil
	}
	_, err = l.svcCtx.MenuModel.Insert(l.ctx, &models.SysMenu{
		ParentId:    in.ParentId,
		Name:        in.Name,
		MenuType:    in.MenuType,
		Method:      in.Method,
		Path:        in.Path,
		Component:   in.Component,
		Perms:       in.Perms,
		Icon:        in.Icon,
		Sort:        in.Sort,
		Visible:     in.Visible,
		Status:      in.Status,
		CreateTimes: time.Now().UnixMilli(),
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: &common.RespBase{
			Code: 200,
			Msg:  "Menu created successfully",
		},
	}, nil
}
