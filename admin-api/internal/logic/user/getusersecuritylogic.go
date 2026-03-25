// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSecurityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityLogic {
	return &GetUserSecurityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSecurityLogic) GetUserSecurity(req *types.GetUserSecurityReq) (resp *types.GetUserSecurityResp, err error) {
	// todo: add your logic here and delete this line

	return
}
