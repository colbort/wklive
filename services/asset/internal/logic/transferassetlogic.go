package logic

import (
	"context"
	"fmt"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TransferAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferAssetLogic {
	return &TransferAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 钱包划转
func (l *TransferAssetLogic) TransferAsset(in *asset.TransferAssetReq) (*asset.TransferAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("TransferAsset parse amount failed, tenantId=%d userId=%d fromWalletType=%d toWalletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.FromWalletType, in.ToWalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
		err := fmt.Errorf("amount must be positive")
		l.Errorf("TransferAsset validate amount failed, tenantId=%d userId=%d fromWalletType=%d toWalletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.FromWalletType, in.ToWalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	ts := utils.NowMillis()
	var (
		beforeFrom *models.TUserAsset
		beforeTo   *models.TUserAsset
		afterFrom  *models.TUserAsset
		afterTo    *models.TUserAsset
	)
	sceneType := assetSceneType(in.SceneType)
	bizType := assetBizType(in.BizType)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

		beforeFrom, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin)
		if err != nil {
			return err
		}

		beforeTo, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin)
		if err != nil && err != models.ErrNotFound {
			return err
		}

		if ok, err := userAssetModel.SubAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, amount, ts); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("insufficient available balance")
		}

		if beforeTo == nil {
			_, err = userAssetModel.Insert(ctx, &models.TUserAsset{
				TenantId:        in.TenantId,
				UserId:          in.UserId,
				WalletType:      int64(in.ToWalletType),
				Coin:            in.Coin,
				TotalAmount:     amount,
				AvailableAmount: amount,
				Status:          1,
				Version:         1,
				Remark:          in.Remark,
				CreateTimes:     ts,
				UpdateTimes:     ts,
			})
			if err != nil {
				return err
			}
		} else {
			if _, err := userAssetModel.AddAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, amount, ts); err != nil {
				return err
			}
		}

		afterFrom, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin)
		if err != nil {
			return err
		}
		afterTo, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin)
		if err != nil {
			return err
		}

		flowOut := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_OUT, amount, beforeFrom, afterFrom, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flowOut); err != nil {
			return err
		}

		flowIn := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_IN, amount, beforeTo, afterTo, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flowIn); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("TransferAsset transaction failed, tenantId=%d userId=%d fromWalletType=%d toWalletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.FromWalletType, in.ToWalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return &asset.TransferAssetResp{Base: helper.OkResp(), FromAsset: toUserAssetProto(afterFrom), ToAsset: toUserAssetProto(afterTo)}, nil
}
