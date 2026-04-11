package logic

import (
	"context"

	"wklive/common/pageutil"
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
	startTime := int64(0)
	endTime := int64(0)
	if in.TimeRange != nil {
		startTime = in.TimeRange.StartTime
		endTime = in.TimeRange.EndTime
	}

	items, total, err := l.svcCtx.AssetFlowModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetBizType(in.BizType), assetSceneType(in.SceneType), "", startTime, endTime, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	resp := &asset.ListMyAssetFlowsResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID)}

	for _, item := range items {
		resp.Data = append(resp.Data, toAssetFlowProto(item))
	}

	return resp, nil
}
