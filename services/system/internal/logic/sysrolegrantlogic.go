package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantLogic {
	return &SysRoleGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleGrantLogic) SysRoleGrant(in *system.SysRoleGrantReq) (*system.SimpleResp, error) {
	// todo: add your logic here and delete this line

	return &system.SimpleResp{}, nil
}
