package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUnfreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUnfreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnfreezeAssetLogic {
	return &AdminUnfreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台解冻资产
func (l *AdminUnfreezeAssetLogic) AdminUnfreezeAsset(in *asset.AdminUnfreezeAssetReq) (*asset.AdminChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.AdminChangeAssetResp{}, nil
}
