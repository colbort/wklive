package applogic

import (
	"context"
	"strings"
	"wklive/services/asset/internal/logic/helpers"

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
		// 资产明细必须始终返回。某个币种暂时没有 USDT 行情时，只应影响
		// 汇总折算，不能导致该币种从用户资产列表中消失。
		resp.Data.Assets = append(resp.Data.Assets, helpers.ToUserAssetProto(item))

		coin := strings.ToUpper(strings.TrimSpace(item.Coin))
		exchangeRate := decimal.NewFromInt(1)
		if coin != "USDT" {
			coinConfig, configErr := l.svcCtx.AssetCoinConfigModel.FindEnabledByWalletCoin(
				l.ctx,
				tenantId,
				item.WalletType,
				coin,
			)
			if configErr != nil {
				l.Errorf(
					"GetMyAssetSummary get coin config failed: tenantId=%d, userId=%d, walletType=%d, coin=%s, err=%v",
					tenantId,
					userId,
					item.WalletType,
					coin,
					configErr,
				)
				continue
			}

			exchangeRate, err = assetUSDTAccountingRate(l.ctx, coin, coinConfig.CoinType, l.svcCtx.LastMarketPrice)
			if err != nil {
				l.Errorf(
					"GetMyAssetSummary get exchange rate failed: tenantId=%d, userId=%d, walletType=%d, coin=%s, coinType=%d, err=%v",
					tenantId,
					userId,
					item.WalletType,
					coin,
					coinConfig.CoinType,
					err,
				)
				continue
			}
		}

		totalAsset = totalAsset.Add(item.TotalAmount.Mul(exchangeRate))
		totalAvailable = totalAvailable.Add(item.AvailableAmount.Mul(exchangeRate))
		totalFrozen = totalFrozen.Add(item.FrozenAmount.Mul(exchangeRate))
		totalLocked = totalLocked.Add(item.LockedAmount.Mul(exchangeRate))
	}

	resp.Data.TotalAssetUsdt = conv.FloatString(totalAsset)
	resp.Data.TotalAvailableUsdt = conv.FloatString(totalAvailable)
	resp.Data.TotalFrozenUsdt = conv.FloatString(totalFrozen)
	resp.Data.TotalLockedUsdt = conv.FloatString(totalLocked)

	return resp, nil
}
