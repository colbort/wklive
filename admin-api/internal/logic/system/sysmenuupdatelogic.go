// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuUpdateLogic {
	return &SysMenuUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysMenuUpdateLogic) SysMenuUpdate(req *types.SysMenuUpdateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysMenuUpdate(l.ctx, &system.SysMenuUpdateReq{
		Id:        req.Id,
		ParentId:  req.ParentId,
		Name:      req.Name,
		MenuType:  req.MenuType,
		Method:    req.Method,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Visible:   req.Visible,
		Status:    req.Status,
		Perms:     req.Perms,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}

	return
}
