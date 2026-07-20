package logic

import (
	"context"
	"errors"

	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

// assetFreezeError distinguishes an explicit Asset rejection from an RPC
// outcome that is unknown to Trade. Only an explicit response is safe to map
// to ORDER_STATUS_REJECTED.
type assetFreezeError struct {
	err        error
	definitive bool
}

func (e *assetFreezeError) Error() string { return e.err.Error() }
func (e *assetFreezeError) Unwrap() error { return e.err }

func isDefinitiveAssetFreezeError(err error) bool {
	var target *assetFreezeError
	return errors.As(err, &target) && target.definitive
}

func freezeOrderAsset(
	svcCtx *svc.ServiceContext,
	ctx context.Context,
	order *models.TTradeOrder,
	symbol *models.TTradeSymbol,
	frozenAsset string,
	frozenAmount decimal.Decimal,
) (string, error) {
	if order == nil || symbol == nil || frozenAsset == "" || !frozenAmount.IsPositive() {
		return "", nil
	}

	resp, err := svcCtx.AssetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: walletTypeForProduct(trade.ProductType(order.ProductType)),
		Coin:       frozenAsset,
		Amount:     frozenAmount.String(),
		BizType:    asset.BizType_BIZ_TYPE_TRADE,
		SceneType:  asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     "trade place order freeze",
	})
	if err != nil {
		return "", &assetFreezeError{err: err}
	}
	if resp == nil || resp.Base == nil {
		return "", &assetFreezeError{err: i18n.StatusError(ctx, i18n.InternalServerError)}
	}
	if resp.Base.Code != 200 {
		return "", &assetFreezeError{err: i18n.StatusError(ctx, resp.Base.Code), definitive: true}
	}

	return resp.GetData().GetFreezeNo(), nil
}

func unfreezeOrderAsset(
	svcCtx *svc.ServiceContext,
	ctx context.Context,
	order *models.TTradeOrder,
	freezeNo string,
	amount decimal.Decimal,
	reason string,
) error {
	if order == nil || freezeNo == "" || !amount.IsPositive() {
		return nil
	}

	resp, err := svcCtx.AssetClient.UnfreezeAsset(ctx, &asset.UnfreezeAssetReq{
		TenantId:  order.TenantId,
		FreezeNo:  freezeNo,
		Amount:    amount.String(),
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
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}

	return nil
}

func unfreezeRemainingOrderAsset(svcCtx *svc.ServiceContext, ctx context.Context, order *models.TTradeOrder, reason string) error {
	if order == nil || order.OrderNo == "" {
		return nil
	}

	resp, err := svcCtx.AssetClient.UnfreezeAssetByBizNo(ctx, &asset.UnfreezeAssetByBizNoReq{
		TenantId:      order.TenantId,
		TargetBizType: asset.BizType_BIZ_TYPE_TRADE,
		TargetBizNo:   order.OrderNo,
		BizType:       asset.BizType_BIZ_TYPE_TRADE,
		SceneType:     asset.SceneType_SCENE_TYPE_CANCEL_ORDER,
		BizId:         order.Id,
		BizNo:         order.OrderNo,
		Remark:        reason,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(ctx, resp.Base.Code)
	}

	return nil
}
