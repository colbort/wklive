package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRolesLogic {
	return &AssignUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分配租户角色
func (l *AssignUserRolesLogic) AssignUserRoles(in *tenant.AssignUserRolesReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
