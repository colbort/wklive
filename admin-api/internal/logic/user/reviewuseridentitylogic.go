// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewUserIdentityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewUserIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewUserIdentityLogic {
	return &ReviewUserIdentityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewUserIdentityLogic) ReviewUserIdentity(req *types.ReviewUserIdentityReq) (resp *types.ReviewUserIdentityResp, err error) {
	return logicutil.Proxy[types.ReviewUserIdentityResp](l.ctx, req, l.svcCtx.UserCli.ReviewUserIdentity)
}
