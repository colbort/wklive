package logic

import (
	"context"

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

type DeductLockedAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductLockedAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductLockedAssetLogic {
	return &DeductLockedAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减锁仓余额
func (l *DeductLockedAssetLogic) DeductLockedAsset(in *asset.DeductLockedAssetReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("DeductLockedAsset parse amount failed, tenantId=%d lockNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("DeductLockedAsset validate amount failed, tenantId=%d lockNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	lock, err := l.svcCtx.AssetLockModel.FindOneByLockNo(l.ctx, in.LockNo)
	if err != nil {
		l.Errorf("DeductLockedAsset find lock failed, tenantId=%d lockNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if lock.TenantId != in.TenantId {
		err := i18n.StatusError(l.ctx, i18n.AssetTenantMismatch)
		l.Errorf("DeductLockedAsset tenant mismatch, tenantId=%d lockTenantId=%d lockNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, lock.TenantId, in.LockNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if amount > lock.RemainAmount {
		err := i18n.StatusError(l.ctx, i18n.DeductAmountExceedsLocked)
		l.Errorf("DeductLockedAsset amount exceeds locked amount, tenantId=%d lockNo=%s amount=%s remainAmount=%v bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, lock.RemainAmount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetLockModel := models.NewTAssetLockModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.DeductLockedAmount(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.DeductLockedFailed)
		}

		ok, err = assetLockModel.UpdateDeduct(ctx, lock.LockNo, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.LockRecordUpdateFailed)
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK_DEDUCT, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("DeductLockedAsset transaction failed, tenantId=%d lockNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.BizNo, Asset: toUserAssetProto(after)}}, nil
}
