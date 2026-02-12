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

type SysUserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDeleteLogic {
	return &SysUserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysUserDeleteLogic) SysUserDelete() (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysUserDelete(l.ctx, &system.SysUserDeleteReq{
		Id: 0,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}, nil
}
