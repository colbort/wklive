package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserAssetDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserAssetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAssetDetailLogic {
	return &GetUserAssetDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户资产详情
func (l *GetUserAssetDetailLogic) GetUserAssetDetail(in *asset.GetUserAssetDetailReq) (*asset.GetUserAssetDetailResp, error) {
	// todo: add your logic here and delete this line

	return &asset.GetUserAssetDetailResp{}, nil
}
