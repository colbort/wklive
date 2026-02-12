package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuUpdateLogic {
	return &SysMenuUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysMenuUpdateLogic) SysMenuUpdate(in *system.SysMenuUpdateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "菜单不存在",
		}, nil
	}

	var data models.SysMenu
	_ = copier.Copy(&data, one)
	_ = copier.Copy(&data, in)

	err = l.svcCtx.MenuModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "更新成功",
	}, nil
}
