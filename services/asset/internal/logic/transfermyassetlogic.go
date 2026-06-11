package logic

import (
	"context"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

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
	if in.FromWalletType == common.WalletType_WALLET_TYPE_UNKNOWN || in.ToWalletType == common.WalletType_WALLET_TYPE_UNKNOWN {
		return nil, i18n.StatusError(l.ctx, i18n.WalletTypeRequired)
	}
	if in.FromWalletType == in.ToWalletType && fromCoin == toCoin {
		return nil, i18n.StatusError(l.ctx, i18n.SameWalletCoinTransferNotNeeded)
	}
	fromAmount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if fromAmount <= 0 {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}
	toAmount, err := l.exchangeTransferAmount(fromCoin, toCoin, fromAmount)
	if err != nil {
		return nil, err
	}

	bizNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "TRANSFER", "")
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

func (l *TransferMyAssetLogic) exchangeTransferAmount(fromCoin, toCoin string, fromAmount float64) (float64, error) {
	if fromCoin == toCoin {
		return fromAmount, nil
	}

	fromRate, err := l.usdtRate(fromCoin)
	if err != nil {
		return 0, err
	}
	toRate, err := l.usdtRate(toCoin)
	if err != nil {
		return 0, err
	}
	if fromRate <= 0 || toRate <= 0 {
		return 0, i18n.StatusError(l.ctx, i18n.InvalidExchangeRate)
	}

	return fromAmount * fromRate / toRate, nil
}

func (l *TransferMyAssetLogic) usdtRate(coin string) (float64, error) {
	if coin == "USDT" {
		return 1, nil
	}
	rate, err := l.svcCtx.LastPrice(l.ctx, coin+"USDT")
	if err != nil {
		l.Errorf("TransferMyAsset get exchange rate failed, coin=%s err=%v", coin, err)
		return 0, err
	}
	return rate, nil
}

func (l *TransferMyAssetLogic) transferAsset(tenantId, userId int64, fromWalletType, toWalletType common.WalletType, fromCoin, toCoin string, fromAmount, toAmount float64, bizNo, remark string) (*asset.TransferMyAssetResp, error) {
	ts := utils.NowMillis()
	var (
		beforeFrom *models.TUserAsset
		beforeTo   *models.TUserAsset
		afterFrom  *models.TUserAsset
		afterTo    *models.TUserAsset
	)

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

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

		flowOut := buildAssetFlowRecord(l.svcCtx, ctx, tenantId, userId, int64(fromWalletType), fromCoin, "transfer", "transfer", "transfer", 0, bizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_OUT, fromAmount, beforeFrom, afterFrom, remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flowOut); err != nil {
			return err
		}

		flowIn := buildAssetFlowRecord(l.svcCtx, ctx, tenantId, userId, int64(toWalletType), toCoin, "transfer", "transfer", "transfer", 0, bizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_IN, toAmount, beforeTo, afterTo, remark, ts)
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

	return &asset.TransferMyAssetResp{Base: helper.OkResp(), Data: &asset.TransferMyAssetData{FromAsset: toUserAssetProto(afterFrom), ToAsset: toUserAssetProto(afterTo)}}, nil
}
