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

type SysRoleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleDeleteLogic {
	return &SysRoleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleDeleteLogic) SysRoleDelete(req *types.SysRoleDeleteReq) (resp *types.SimpleResp, err error) {
	result, err := l.svcCtx.SystemCli.SysRoleDelete(l.ctx, &system.SysRoleDeleteReq{
		Id: req.Id,
	})
	return &types.SimpleResp{
		Code: int(result.Code),
		Msg:  result.Msg,
	}, err
}
