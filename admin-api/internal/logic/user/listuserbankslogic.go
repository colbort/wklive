// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserBanksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserBanksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserBanksLogic {
	return &ListUserBanksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserBanksLogic) ListUserBanks(req *types.ListUserBanksReq) (resp *types.ListUserBanksResp, err error) {
	return logicutil.Proxy[types.ListUserBanksResp](l.ctx, req, l.svcCtx.UserCli.ListUserBanks)
}
