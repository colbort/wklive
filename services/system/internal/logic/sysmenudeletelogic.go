package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuDeleteLogic {
	return &SysMenuDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete a menu
func (l *SysMenuDeleteLogic) SysMenuDelete(in *system.SysMenuDeleteReq) (*system.SimpleResp, error) {
	// todo: add your logic here and delete this line

	return &system.SimpleResp{}, nil
}
