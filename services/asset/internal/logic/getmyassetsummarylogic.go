package logic

import (
	"context"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"

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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	list, err := l.svcCtx.UserAssetModel.FindAll(l.ctx, models.UserAssetPageFilter{
		TenantId: tenantId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	totalAsset := decimal.Zero
	totalAvailable := decimal.Zero
	totalFrozen := decimal.Zero
	totalLocked := decimal.Zero
	resp := &asset.GetMyAssetSummaryResp{Base: helper.OkResp(), Data: &asset.UserAssetSummary{TenantId: tenantId, UserId: userId}}
	for _, item := range list {
		// 总资产、可用资产、冻结资产、锁定资产，单位都是USDT
		if item.Coin == "USDT" {
			totalAsset = totalAsset.Add(item.TotalAmount)
			totalAvailable = totalAvailable.Add(item.AvailableAmount)
			totalFrozen = totalFrozen.Add(item.FrozenAmount)
			totalLocked = totalLocked.Add(item.LockedAmount)
		} else {
			// 其他币种需要换算成USDT
			exchangeRate, err := l.svcCtx.LastPrice(l.ctx, item.Coin+"USDT")
			if err != nil {
				logx.Errorf("GetExchangeRate error: tenantId=%d, coin=%s, err=%v", tenantId, item.Coin, err)
				continue
			}
			totalAsset = totalAsset.Add(item.TotalAmount.Mul(exchangeRate))
			totalAvailable = totalAvailable.Add(item.AvailableAmount.Mul(exchangeRate))
			totalFrozen = totalFrozen.Add(item.FrozenAmount.Mul(exchangeRate))
			totalLocked = totalLocked.Add(item.LockedAmount.Mul(exchangeRate))
		}
		resp.Data.Assets = append(resp.Data.Assets, toUserAssetProto(item))
	}

	resp.Data.TotalAssetUsdt = conv.FloatString(totalAsset)
	resp.Data.TotalAvailableUsdt = conv.FloatString(totalAvailable)
	resp.Data.TotalFrozenUsdt = conv.FloatString(totalFrozen)
	resp.Data.TotalLockedUsdt = conv.FloatString(totalLocked)

	return resp, nil
}
