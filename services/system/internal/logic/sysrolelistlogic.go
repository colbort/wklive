package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleListLogic {
	return &SysRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色
func (l *SysRoleListLogic) SysRoleList(in *system.SysRoleListReq) (*system.SysRoleListResp, error) {
	// todo: add your logic here and delete this line

	return &system.SysRoleListResp{}, nil
}
