// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleGrantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleGrantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantDetailLogic {
	return &SysRoleGrantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleGrantDetailLogic) SysRoleGrantDetail(req *types.SysRoleGrantDetailReq) (resp *types.SysRoleGrantDetailResp, err error) {
	return logicutil.Proxy[types.SysRoleGrantDetailResp](l.ctx, req, l.svcCtx.SystemCli.SysRoleGrantDetail)
}
