// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"wklive/app-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdentityLogic {
	return &GetIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetIdentityLogic) GetIdentity() (resp *types.GetIdentityResp, err error) {
	return logicutil.Proxy[types.GetIdentityResp](l.ctx, nil, l.svcCtx.UserCli.GetIdentity)
}
