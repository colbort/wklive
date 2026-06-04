// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

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

func (l *SysMenuTreeLogic) SysMenuTree(req *types.SysMenuTreeReq) (resp *types.SysMenuTreeResp, err error) {
	return logicutil.Proxy[types.SysMenuTreeResp](l.ctx, req, l.svcCtx.SystemCli.GetMenuTree)
}
