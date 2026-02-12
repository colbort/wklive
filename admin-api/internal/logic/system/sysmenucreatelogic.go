// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
