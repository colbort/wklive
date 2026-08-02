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

type CreatePlatformBackstopPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePlatformBackstopPolicyLogic {
	return &CreatePlatformBackstopPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePlatformBackstopPolicyLogic) CreatePlatformBackstopPolicy(req *types.CreatePlatformBackstopPolicyReq) (resp *types.PlatformBackstopPolicyResp, err error) {
	return logicutil.Proxy[types.PlatformBackstopPolicyResp](
		l.ctx, req, l.svcCtx.AssetCli.CreatePlatformBackstopPolicy,
	)
}
