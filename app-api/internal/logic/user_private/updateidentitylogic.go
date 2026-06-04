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

type UpdateIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIdentityLogic {
	return &UpdateIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateIdentityLogic) UpdateIdentity(req *types.UpdateIdentityReq) (resp *types.UpdateIdentityResp, err error) {
	return logicutil.Proxy[types.UpdateIdentityResp](l.ctx, req, l.svcCtx.UserCli.UpdateIdentity)
}
