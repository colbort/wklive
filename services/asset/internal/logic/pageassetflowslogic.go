package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetFlowsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFlowsLogic {
	return &PageAssetFlowsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资产流水
func (l *PageAssetFlowsLogic) PageAssetFlows(in *asset.PageAssetFlowsReq) (*asset.PageAssetFlowsResp, error) {
	startTime := int64(0)
	endTime := int64(0)
	if in.TimeRange != nil {
		startTime = in.TimeRange.StartTime
		endTime = in.TimeRange.EndTime
	}

	flows, total, err := l.svcCtx.AssetFlowModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, startTime, endTime, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(flows)) == in.Page.Limit && in.Page.Limit > 0 {
		nextCursor = flows[len(flows)-1].Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(flows)) == in.Page.Limit && in.Page.Limit > 0

	resp := &asset.PageAssetFlowsResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}

	for _, item := range flows {
		resp.Data = append(resp.Data, helpers.ToAssetFlowProto(item))
	}
	return resp, nil
}
