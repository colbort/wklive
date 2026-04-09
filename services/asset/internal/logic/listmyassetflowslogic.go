package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyAssetFlowsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyAssetFlowsLogic {
	return &ListMyAssetFlowsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的资产流水
func (l *ListMyAssetFlowsLogic) ListMyAssetFlows(in *asset.ListMyAssetFlowsReq) (*asset.ListMyAssetFlowsResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ListMyAssetFlowsResp{}, nil
}
