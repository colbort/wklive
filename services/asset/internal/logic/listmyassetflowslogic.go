package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
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

	items, total, err := l.svcCtx.AssetFlowModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), "", startTime, endTime, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	resp := &asset.ListMyAssetFlowsResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}

	for _, item := range items {
		resp.Data = append(resp.Data, helpers.ToAssetFlowProto(item))
	}

	return resp, nil
}
