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

type SysUserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserCreateLogic {
	return &SysUserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysUserCreateLogic) SysUserCreate(req *types.SysUserCreateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysUserCreate(l.ctx, &system.SysUserCreateReq{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: req.Password,
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
