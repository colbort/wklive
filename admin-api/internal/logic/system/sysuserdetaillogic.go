// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDetailLogic {
	return &SysUserDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysUserDetailLogic) SysUserDetail() (resp *types.SysUserDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
