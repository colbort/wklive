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

type TransferMyAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransferMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferMyAssetLogic {
	return &TransferMyAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransferMyAssetLogic) TransferMyAsset(req *types.TransferMyAssetReq) (resp *types.TransferMyAssetResp, err error) {
	return logicutil.Proxy[types.TransferMyAssetResp](l.ctx, req, l.svcCtx.AssetCli.TransferMyAsset)
}
