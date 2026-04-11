// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyFreezesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyFreezesLogic {
	return &ListMyFreezesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyFreezesLogic) ListMyFreezes(req *types.ListMyFreezesReq) (resp *types.ListMyFreezesResp, err error) {
	return logicutil.Proxy[types.ListMyFreezesResp](l.ctx, req, l.svcCtx.AssetCli.ListMyFreezes)
}
