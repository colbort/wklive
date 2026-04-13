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
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
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

		if _, err := userAssetModel.AddAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, amount, 0, ts); err != nil {
			return err
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
		return nil, err
	}

	return &asset.TransferAssetResp{Base: helper.OkResp(), FromAsset: toUserAssetProto(afterFrom), ToAsset: toUserAssetProto(afterTo)}, nil
}
