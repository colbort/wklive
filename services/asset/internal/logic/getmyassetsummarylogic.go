package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
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
	list, err := l.svcCtx.UserAssetModel.FindAll(l.ctx, in.TenantId, in.UserId, 0, "", 0)
	if err != nil {
		return nil, err
	}

	totalAsset := 0.0
	totalAvailable := 0.0
	totalFrozen := 0.0
	totalLocked := 0.0
	resp := &asset.GetMyAssetSummaryResp{Base: helper.OkResp(), Data: &asset.UserAssetSummary{TenantId: in.TenantId, UserId: in.UserId}}
	for _, item := range list {
		totalAsset += item.TotalAmount
		totalAvailable += item.AvailableAmount
		totalFrozen += item.FrozenAmount
		totalLocked += item.LockedAmount
		resp.Data.Assets = append(resp.Data.Assets, helpers.ToUserAssetProto(item))
	}

	resp.Data.TotalAssetUsdt = helpers.FormatAmount(totalAsset)
	resp.Data.TotalAvailableUsdt = helpers.FormatAmount(totalAvailable)
	resp.Data.TotalFrozenUsdt = helpers.FormatAmount(totalFrozen)
	resp.Data.TotalLockedUsdt = helpers.FormatAmount(totalLocked)

	return resp, nil
}
