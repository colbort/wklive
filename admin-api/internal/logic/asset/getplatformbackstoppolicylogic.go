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

type GetPlatformBackstopPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformBackstopPolicyLogic {
	return &GetPlatformBackstopPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlatformBackstopPolicyLogic) GetPlatformBackstopPolicy(req *types.GetPlatformBackstopPolicyReq) (resp *types.PlatformBackstopPolicyResp, err error) {
	return logicutil.Proxy[types.PlatformBackstopPolicyResp](
		l.ctx, req, l.svcCtx.AssetCli.GetPlatformBackstopPolicy,
	)
}
