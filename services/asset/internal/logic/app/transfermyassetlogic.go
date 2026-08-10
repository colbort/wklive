package applogic

import (
	"context"
	"errors"
	"strings"
	"wklive/services/asset/internal/logic/helpers"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	assetCoinTypeFiat   = int64(asset.AssetCoinType_ASSET_COIN_TYPE_FIAT)
	assetCoinTypeCrypto = int64(asset.AssetCoinType_ASSET_COIN_TYPE_CRYPTO)
	assetPriceScaleMax  = int64(18)
)

var (
	errExchangeRateUnavailable = errors.New("exchange rate unavailable")
	usdUSDTAccountingRate      = decimal.NewFromInt(1)
)

type marketQuoteReader func(ctx context.Context, categoryCode, market, symbol string) (decimal.Decimal, error)

type TransferMyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferMyAssetLogic {
	return &TransferMyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 我的账户划转
func (l *TransferMyAssetLogic) TransferMyAsset(in *asset.TransferMyAssetReq) (*asset.TransferMyAssetResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}

	fromCoin := strings.ToUpper(strings.TrimSpace(in.FromCoin))
	toCoin := strings.ToUpper(strings.TrimSpace(in.ToCoin))
	if fromCoin == "" || toCoin == "" {
		return nil, i18n.StatusError(l.ctx, i18n.TransferCoinRequired)
	}
	if !validTransferWalletType(in.FromWalletType) || !validTransferWalletType(in.ToWalletType) {
		return nil, i18n.StatusError(l.ctx, i18n.WalletTypeRequired)
	}
	if in.FromWalletType == in.ToWalletType && fromCoin == toCoin {
		return nil, i18n.StatusError(l.ctx, i18n.SameWalletCoinTransferNotNeeded)
	}
	fromAmount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		return nil, err
	}
	if !fromAmount.IsPositive() {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}
	fromConfig, err := l.findTransferCoinConfig(tenantId, in.FromWalletType, fromCoin)
	if err != nil {
		return nil, err
	}
	toConfig, err := l.findTransferCoinConfig(tenantId, in.ToWalletType, toCoin)
	if err != nil {
		return nil, err
	}

	toAmount, err := exchangeTransferAmount(
		l.ctx,
		fromConfig,
		toConfig,
		fromAmount,
		l.svcCtx.LastMarketPrice,
	)
	if err != nil {
		l.Errorf(
			"TransferMyAsset resolve exchange rate failed, tenantId=%d fromWalletType=%d toWalletType=%d fromCoin=%s toCoin=%s fromCoinType=%d toCoinType=%d err=%v",
			tenantId,
			in.FromWalletType,
			in.ToWalletType,
			fromCoin,
			toCoin,
			fromConfig.CoinType,
			toConfig.CoinType,
			err,
		)
		return nil, i18n.StatusError(l.ctx, i18n.InvalidExchangeRate)
	}

	bizNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "TRANSFER", "")
	if err != nil {
		return nil, err
	}
	result, err := l.transferAsset(tenantId, userId, in.FromWalletType, in.ToWalletType, fromCoin, toCoin, fromAmount, toAmount, bizNo, in.Remark)
	if err != nil {
		return nil, err
	}

	return &asset.TransferMyAssetResp{
		Base: result.GetBase(),
		Data: &asset.TransferMyAssetData{
			FromAsset: result.GetData().GetFromAsset(),
			ToAsset:   result.GetData().GetToAsset(),
		},
	}, nil
}

func validTransferWalletType(walletType common.WalletType) bool {
	switch walletType {
	case common.WalletType_WALLET_TYPE_SPOT,
		common.WalletType_WALLET_TYPE_FUNDING,
		common.WalletType_WALLET_TYPE_CONTRACT,
		common.WalletType_WALLET_TYPE_EARN,
		common.WalletType_WALLET_TYPE_OPTION:
		return true
	default:
		return false
	}
}

func (l *TransferMyAssetLogic) findTransferCoinConfig(tenantId int64, walletType common.WalletType, coin string) (*models.TAssetCoinConfig, error) {
	config, err := l.svcCtx.AssetCoinConfigModel.FindTransferEnabled(l.ctx, tenantId, int64(walletType), coin)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, i18n.StatusError(l.ctx, i18n.AssetCoinConfigNotFound)
		}
		return nil, err
	}
	if config.CoinType != assetCoinTypeFiat && config.CoinType != assetCoinTypeCrypto {
		return nil, i18n.StatusError(l.ctx, i18n.AssetCoinConfigNotFound)
	}
	return config, nil
}

func exchangeTransferAmount(
	ctx context.Context,
	fromConfig, toConfig *models.TAssetCoinConfig,
	fromAmount decimal.Decimal,
	quote marketQuoteReader,
) (decimal.Decimal, error) {
	if fromConfig == nil || toConfig == nil || quote == nil {
		return decimal.Zero, errExchangeRateUnavailable
	}
	if fromConfig.Coin == toConfig.Coin {
		return fromAmount, nil
	}

	rate, err := resolveExchangeRate(ctx, fromConfig.Coin, fromConfig.CoinType, toConfig.Coin, toConfig.CoinType, quote)
	if err != nil {
		return decimal.Zero, err
	}
	toAmount := roundAssetAmount(fromAmount.Mul(rate), toConfig.DecimalPlaces)
	if !toAmount.IsPositive() {
		return decimal.Zero, errExchangeRateUnavailable
	}
	return toAmount, nil
}

func roundAssetAmount(amount decimal.Decimal, decimalPlaces int64) decimal.Decimal {
	if decimalPlaces < 0 {
		decimalPlaces = 0
	}
	if decimalPlaces > assetPriceScaleMax {
		decimalPlaces = assetPriceScaleMax
	}
	return amount.Round(int32(decimalPlaces))
}

func resolveExchangeRate(
	ctx context.Context,
	fromCoin string,
	fromCoinType int64,
	toCoin string,
	toCoinType int64,
	quote marketQuoteReader,
) (decimal.Decimal, error) {
	fromCoin = strings.ToUpper(strings.TrimSpace(fromCoin))
	toCoin = strings.ToUpper(strings.TrimSpace(toCoin))
	if fromCoin == "" || toCoin == "" {
		return decimal.Zero, errExchangeRateUnavailable
	}
	if fromCoin == toCoin {
		return decimal.NewFromInt(1), nil
	}

	var categoryCode, market string
	switch {
	case fromCoinType == assetCoinTypeFiat && toCoinType == assetCoinTypeFiat:
		categoryCode, market = "forex", "GB"
	case fromCoinType == assetCoinTypeCrypto && toCoinType == assetCoinTypeCrypto:
		categoryCode, market = "crypto", "BA"
	case (fromCoinType == assetCoinTypeFiat && toCoinType == assetCoinTypeCrypto) ||
		(fromCoinType == assetCoinTypeCrypto && toCoinType == assetCoinTypeFiat):
		// Mixed fiat/crypto conversion uses USD and USDT as accounting anchors.
	default:
		return decimal.Zero, errExchangeRateUnavailable
	}

	if categoryCode != "" {
		if rate, err := directOrInverseRate(ctx, categoryCode, market, fromCoin, toCoin, quote); err == nil {
			return rate, nil
		} else if !errors.Is(err, errExchangeRateUnavailable) {
			return decimal.Zero, err
		}
	}

	fromRate, err := assetUSDTAccountingRate(ctx, fromCoin, fromCoinType, quote)
	if err != nil {
		return decimal.Zero, err
	}
	toRate, err := assetUSDTAccountingRate(ctx, toCoin, toCoinType, quote)
	if err != nil {
		return decimal.Zero, err
	}
	if !fromRate.IsPositive() || !toRate.IsPositive() {
		return decimal.Zero, errExchangeRateUnavailable
	}
	return fromRate.Div(toRate), nil
}

func assetUSDTAccountingRate(
	ctx context.Context,
	coin string,
	coinType int64,
	quote marketQuoteReader,
) (decimal.Decimal, error) {
	switch coinType {
	case assetCoinTypeCrypto:
		if coin == "USDT" {
			return decimal.NewFromInt(1), nil
		}
		return directOrInverseRate(ctx, "crypto", "BA", coin, "USDT", quote)
	case assetCoinTypeFiat:
		if coin == "USD" {
			return usdUSDTAccountingRate, nil
		}
		usdPerCoin, err := directOrInverseRate(ctx, "forex", "GB", coin, "USD", quote)
		if err != nil {
			return decimal.Zero, err
		}
		return usdPerCoin.Mul(usdUSDTAccountingRate), nil
	default:
		return decimal.Zero, errExchangeRateUnavailable
	}
}

func directOrInverseRate(
	ctx context.Context,
	categoryCode, market, fromCoin, toCoin string,
	quote marketQuoteReader,
) (decimal.Decimal, error) {
	direct, err := quote(ctx, categoryCode, market, fromCoin+toCoin)
	if err == nil && direct.IsPositive() {
		return direct, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		return decimal.Zero, err
	}

	inverse, inverseErr := quote(ctx, categoryCode, market, toCoin+fromCoin)
	if inverseErr == nil && inverse.IsPositive() {
		return decimal.NewFromInt(1).Div(inverse), nil
	}
	if inverseErr != nil && !errors.Is(inverseErr, redis.Nil) {
		return decimal.Zero, inverseErr
	}
	return decimal.Zero, errExchangeRateUnavailable
}

func (l *TransferMyAssetLogic) transferAsset(tenantId, userId int64, fromWalletType, toWalletType common.WalletType, fromCoin, toCoin string, fromAmount, toAmount decimal.Decimal, bizNo, remark string) (*asset.TransferMyAssetResp, error) {
	ts := utils.NowMillis()
	var (
		beforeFrom *models.TUserAsset
		beforeTo   *models.TUserAsset
		afterFrom  *models.TUserAsset
		afterTo    *models.TUserAsset
	)

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)

		var err error
		beforeFrom, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, tenantId, userId, int64(fromWalletType), fromCoin)
		if err != nil {
			return err
		}

		beforeTo, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, tenantId, userId, int64(toWalletType), toCoin)
		if err != nil && err != models.ErrNotFound {
			return err
		}

		if ok, err := userAssetModel.SubAvailableAmount(ctx, tenantId, userId, int64(fromWalletType), fromCoin, fromAmount, ts); err != nil {
			return err
		} else if !ok {
			return i18n.StatusError(ctx, i18n.InsufficientAvailableBalance)
		}

		if beforeTo == nil {
			_, err = userAssetModel.Insert(ctx, &models.TUserAsset{
				TenantId:        tenantId,
				UserId:          userId,
				WalletType:      int64(toWalletType),
				Coin:            toCoin,
				TotalAmount:     toAmount,
				AvailableAmount: toAmount,
				Enabled:         1,
				Version:         1,
				Remark:          remark,
				CreateTimes:     ts,
				UpdateTimes:     ts,
			})
			if err != nil {
				return err
			}
		} else if _, err := userAssetModel.AddAvailableAmount(ctx, tenantId, userId, int64(toWalletType), toCoin, toAmount, ts); err != nil {
			return err
		}

		afterFrom, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, tenantId, userId, int64(fromWalletType), fromCoin)
		if err != nil {
			return err
		}
		afterTo, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, tenantId, userId, int64(toWalletType), toCoin)
		if err != nil {
			return err
		}

		flowOut := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, tenantId, userId, int64(fromWalletType), fromCoin, "transfer", "transfer", "transfer", 0, bizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_OUT, fromAmount, beforeFrom, afterFrom, remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flowOut); err != nil {
			return err
		}

		flowIn := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, tenantId, userId, int64(toWalletType), toCoin, "transfer", "transfer", "transfer", 0, bizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_IN, toAmount, beforeTo, afterTo, remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flowIn); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("TransferMyAsset transaction failed, tenantId=%d userId=%d fromWalletType=%d toWalletType=%d fromCoin=%s toCoin=%s fromAmount=%v toAmount=%v bizNo=%s err=%v",
			tenantId, userId, fromWalletType, toWalletType, fromCoin, toCoin, fromAmount, toAmount, bizNo, err)
		return nil, err
	}

	return &asset.TransferMyAssetResp{Base: helper.OkResp(), Data: &asset.TransferMyAssetData{FromAsset: helpers.ToUserAssetProto(afterFrom), ToAsset: helpers.ToUserAssetProto(afterTo)}}, nil
}
