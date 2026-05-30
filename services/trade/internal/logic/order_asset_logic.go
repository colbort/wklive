package logic

import (
	"context"
	"fmt"

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
