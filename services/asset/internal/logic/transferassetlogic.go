package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferAssetLogic {
	return &TransferAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 钱包划转
func (l *TransferAssetLogic) TransferAsset(in *asset.TransferAssetReq) (*asset.TransferAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.TransferAssetResp{}, nil
}
