package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuListLogic {
	return &SysMenuListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取菜单列表
func (l *SysMenuListLogic) SysMenuList(in *system.SysMenuListReq) (*system.SysMenuListResp, error) {
	// todo: add your logic here and delete this line

	return &system.SysMenuListResp{}, nil
}
