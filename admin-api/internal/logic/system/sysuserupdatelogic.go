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

type SysUserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserUpdateLogic {
	return &SysUserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysUserUpdateLogic) SysUserUpdate(req *types.SysUserUpdateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysUserUpdate(l.ctx, &system.SysUserUpdateReq{
		Id:       req.Id,
		Nickname: req.Nickname,
		Status:   req.Status,
		RoleIds:  req.RoleIds,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}, nil
}
