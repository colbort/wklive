package assetlogic

import (
	"context"
	"wklive/services/asset/internal/logic/helpers"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type LockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockAssetLogic {
	return &LockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 锁仓
func (l *LockAssetLogic) LockAsset(in *asset.LockAssetReq) (*asset.LockAssetResp, error) {
	amount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		l.Errorf("LockAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizNo, err)
		return nil, err
	}
	if !amount.IsPositive() {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("LockAsset validate amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var (
		after *models.TUserAsset
		lock  *models.TAssetLock
	)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetLockModel := models.NewTAssetLockModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotentModel := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		if in.BizNo != "" {
			done, err := helpers.PrepareAssetIdempotent(ctx, idempotentModel, in.TenantId, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, ts)
			if err != nil {
				return err
			}
			if done {
				lock, err = assetLockModel.FindOneByTenantBizNo(ctx, in.TenantId, helpers.AssetBizType(in.BizType), in.BizNo)
				if err != nil {
					return err
				}
				if lock.UserId != in.UserId || lock.WalletType != int64(in.WalletType) || lock.Coin != in.Coin || !lock.Amount.Equal(amount) {
					return i18n.StatusError(ctx, i18n.ParamError)
				}
				after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
				return err
			}
		}

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.LockAmount(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.InsufficientAvailableBalance)
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		lock = helpers.BuildAssetLockRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.StartTime, in.EndTime, ts)
		if _, err := assetLockModel.Insert(ctx, lock); err != nil {
			return err
		}

		flow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetSceneType(in.SceneType), helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		if in.BizNo != "" {
			if err := helpers.CompleteAssetIdempotent(ctx, idempotentModel, in.TenantId, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, ts); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.Errorf("LockAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizNo, err)
		return nil, err
	}

	return &asset.LockAssetResp{Base: helper.OkResp(), Data: &asset.LockAssetData{LockNo: lock.LockNo, Asset: helpers.ToUserAssetProto(after)}}, nil
}
