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

type GetUserAssetDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserAssetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAssetDetailLogic {
	return &GetUserAssetDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserAssetDetailLogic) GetUserAssetDetail(req *types.GetUserAssetDetailReq) (resp *types.GetUserAssetDetailResp, err error) {
	return logicutil.Proxy[types.GetUserAssetDetailResp](l.ctx, req, l.svcCtx.AssetCli.GetUserAssetDetail)
}
