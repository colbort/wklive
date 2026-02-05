// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleGrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleGrantLogic {
	return &RoleGrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleGrantLogic) RoleGrant(req *types.SysRoleGrantReq) (resp *types.SimpleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
