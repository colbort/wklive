// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserIdentitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserIdentitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserIdentitiesLogic {
	return &ListUserIdentitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserIdentitiesLogic) ListUserIdentities(req *types.ListUserIdentitiesReq) (resp *types.ListUserIdentitiesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
