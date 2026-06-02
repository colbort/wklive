package logic

import (
	"context"
	"errors"
	"fmt"
	"math"

	"wklive/common/conv"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

func freezeOrderAsset(
	svcCtx *svc.ServiceContext,
	ctx context.Context,
	order *models.TTradeOrder,
	symbol *models.TTradeSymbol,
	frozenAsset string,
	frozenAmount float64,
) (string, error) {
	if order == nil || symbol == nil || frozenAsset == "" || frozenAmount <= 0 {
		return "", nil
	}

	resp, err := svcCtx.AssetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: walletTypeForMarket(trade.MarketType(order.MarketType)),
		Coin:       frozenAsset,
		Amount:     fmt.Sprintf("%v", frozenAmount),
		BizType:    asset.BizType_BIZ_TYPE_TRADE,
		SceneType:  asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     "trade place order freeze",
	})
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Base == nil {
		return "", fmt.Errorf("asset freeze returned empty response")
	}
	if resp.Base.Code != 200 {
		return "", fmt.Errorf("asset freeze failed: %s", resp.Base.Msg)
	}

	return resp.FreezeNo, nil
}

func unfreezeOrderAsset(
	svcCtx *svc.ServiceContext,
	ctx context.Context,
	order *models.TTradeOrder,
	freezeNo string,
	amount float64,
	reason string,
) error {
	if order == nil || freezeNo == "" || amount <= 0 {
		return nil
	}

	resp, err := svcCtx.AssetClient.UnfreezeAsset(ctx, &asset.UnfreezeAssetReq{
		TenantId:  order.TenantId,
		FreezeNo:  freezeNo,
		Amount:    fmt.Sprintf("%v", amount),
		BizType:   asset.BizType_BIZ_TYPE_TRADE,
		SceneType: asset.SceneType_SCENE_TYPE_CANCEL_ORDER,
		BizId:     order.Id,
		BizNo:     order.OrderNo,
		Remark:    reason,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return fmt.Errorf("asset unfreeze returned empty response")
	}
	if resp.Base.Code != 200 {
		return fmt.Errorf("asset unfreeze failed: %s", resp.Base.Msg)
	}

	return nil
}

func unfreezeRemainingOrderAsset(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder, reason string) error {
	if order == nil {
		return nil
	}
	ext, err := parseOrderAssetExt(conv.NullStringValue(order.BizExt))
	if err != nil {
		return err
	}
	if ext.FreezeNo == "" {
		return nil
	}
	amount, err := remainingFrozenAmount(svcCtx, ctx, order)
	if err != nil {
		return err
	}
	return unfreezeOrderAsset(svcCtx, ctx, order, ext.FreezeNo, amount, reason)
}

func remainingFrozenAmount(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder) (float64, error) {
	if order.MarketType == int64(trade.MarketType_MARKET_TYPE_SPOT) {
		spot, err := svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(ctx, order.TenantId, order.Id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return 0, nil
			}
			return 0, err
		}
		if order.Side == int64(trade.TradeSide_TRADE_SIDE_SELL) {
			return math.Max(spot.FrozenAmount-order.FilledQty, 0), nil
		}
		return math.Max(spot.FrozenAmount-order.FilledAmount, 0), nil
	}

	contract, err := svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(ctx, order.TenantId, order.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return math.Max(contract.MarginAmount-order.FilledAmount, 0), nil
}
