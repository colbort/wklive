package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAssetLogic {
	return &GetMyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的单个币种资产
func (l *GetMyAssetLogic) GetMyAsset(in *asset.GetMyAssetReq) (*asset.GetMyAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.GetMyAssetResp{}, nil
}
