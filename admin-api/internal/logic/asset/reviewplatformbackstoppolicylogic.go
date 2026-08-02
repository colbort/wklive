// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewPlatformBackstopPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewPlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewPlatformBackstopPolicyLogic {
	return &ReviewPlatformBackstopPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewPlatformBackstopPolicyLogic) ReviewPlatformBackstopPolicy(req *types.ReviewPlatformBackstopPolicyReq) (resp *types.PlatformBackstopPolicyResp, err error) {
	return logicutil.Proxy[types.PlatformBackstopPolicyResp](
		l.ctx, req, l.svcCtx.AssetCli.ReviewPlatformBackstopPolicy,
	)
}
