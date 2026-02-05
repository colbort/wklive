package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDetailLogic {
	return &SysUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserDetailLogic) SysUserDetail(in *system.SysUserDetailReq) (*system.SysUserDetailResp, error) {
	// todo: add your logic here and delete this line

	return &system.SysUserDetailResp{}, nil
}
