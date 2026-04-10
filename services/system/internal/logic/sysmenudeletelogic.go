package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuDeleteLogic {
	return &SysMenuDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete a menu
func (l *SysMenuDeleteLogic) SysMenuDelete(in *system.SysMenuDeleteReq) (*system.RespBase, error) {
	menu, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, "Menu not found"),
		}, nil
	}
	err = l.svcCtx.MenuModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
