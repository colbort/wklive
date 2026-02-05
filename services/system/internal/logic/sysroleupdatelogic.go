package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleUpdateLogic) SysRoleUpdate(in *system.SysRoleUpdateReq) (*system.SimpleResp, error) {
	// todo: add your logic here and delete this line

	return &system.SimpleResp{}, nil
}
