package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyAssetSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyAssetSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAssetSummaryLogic {
	return &GetMyAssetSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的资产汇总
func (l *GetMyAssetSummaryLogic) GetMyAssetSummary(in *asset.GetMyAssetSummaryReq) (*asset.GetMyAssetSummaryResp, error) {
	// todo: add your logic here and delete this line

	return &asset.GetMyAssetSummaryResp{}, nil
}
