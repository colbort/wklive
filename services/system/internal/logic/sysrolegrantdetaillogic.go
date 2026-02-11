package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleGrantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleGrantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantDetailLogic {
	return &SysRoleGrantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取角色授权详情
func (l *SysRoleGrantDetailLogic) SysRoleGrantDetail(in *system.SysRoleGrantDetailReq) (*system.SysRoleGrantDetailResp, error) {
	// todo: add your logic here and delete this line

	return &system.SysRoleGrantDetailResp{}, nil
}
