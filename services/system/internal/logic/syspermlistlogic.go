package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysPermListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysPermListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysPermListLogic {
	return &SysPermListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取权限列表
func (l *SysPermListLogic) SysPermList(in *system.Empty) (*system.SysPermListResp, error) {
	// todo: add your logic here and delete this line

	return &system.SysPermListResp{}, nil
}
