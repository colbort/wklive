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

type ListPlatformBackstopPoliciesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPlatformBackstopPoliciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPlatformBackstopPoliciesLogic {
	return &ListPlatformBackstopPoliciesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPlatformBackstopPoliciesLogic) ListPlatformBackstopPolicies(req *types.ListPlatformBackstopPoliciesReq) (resp *types.ListPlatformBackstopPoliciesResp, err error) {
	return logicutil.Proxy[types.ListPlatformBackstopPoliciesResp](
		l.ctx, req, l.svcCtx.AssetCli.ListPlatformBackstopPolicies,
	)
}
