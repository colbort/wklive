package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

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

	flows, total, err := l.svcCtx.AssetFlowModel.FindPage(l.ctx, models.AssetFlowPageFilter{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		WalletType: int64(in.WalletType),
		Coin:       in.Coin,
		BizType:    assetBizType(in.BizType),
		SceneType:  assetSceneType(in.SceneType),
		BizNo:      in.BizNo,
		StartTime:  startTime,
		EndTime:    endTime,
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(flows) > 0 {
		lastID = flows[len(flows)-1].Id
	}

	resp := &asset.PageAssetFlowsResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(flows), total, lastID)}

	for _, item := range flows {
		resp.Data = append(resp.Data, toAssetFlowProto(item))
	}
	return resp, nil
}
