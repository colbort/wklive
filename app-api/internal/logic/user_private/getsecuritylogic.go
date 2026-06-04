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

type GetSecurityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecurityLogic {
	return &GetSecurityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSecurityLogic) GetSecurity() (resp *types.GetSecurityResp, err error) {
	return logicutil.Proxy[types.GetSecurityResp](l.ctx, nil, l.svcCtx.UserCli.GetSecurity)
}
