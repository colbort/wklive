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

type SysMenuCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuCreateLogic {
	return &SysMenuCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysMenuCreateLogic) SysMenuCreate(req *types.SysMenuCreateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysMenuCreate(l.ctx, &system.SysMenuCreateReq{
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
	return &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}, nil
}
