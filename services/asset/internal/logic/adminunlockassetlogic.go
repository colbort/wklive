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

type AdminUnlockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnlockAssetLogic {
	return &AdminUnlockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台解锁资产
func (l *AdminUnlockAssetLogic) AdminUnlockAsset(in *asset.AdminUnlockAssetReq) (*asset.AdminChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("AdminUnlockAsset parse amount failed, tenantId=%d lockNo=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("AdminUnlockAsset validate amount failed, tenantId=%d lockNo=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizNo, err)
		return nil, err
	}

	lock, err := l.svcCtx.AssetLockModel.FindOneByLockNo(l.ctx, in.LockNo)
	if err != nil {
		l.Errorf("AdminUnlockAsset find lock failed, tenantId=%d lockNo=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizNo, err)
		return nil, err
	}
	if lock.TenantId != in.TenantId {
		err := i18n.StatusError(l.ctx, i18n.AssetTenantMismatch)
		l.Errorf("AdminUnlockAsset tenant mismatch, tenantId=%d lockTenantId=%d lockNo=%s amount=%s bizNo=%s err=%v",
			in.TenantId, lock.TenantId, in.LockNo, in.Amount, in.BizNo, err)
		return nil, err
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, lock.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &asset.AdminChangeAssetResp{Base: base}, nil
	}
	if amount > lock.RemainAmount {
		err := i18n.StatusError(l.ctx, i18n.UnlockAmountExceedsLocked)
		l.Errorf("AdminUnlockAsset amount exceeds locked amount, tenantId=%d lockNo=%s amount=%s remainAmount=%v bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, lock.RemainAmount, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetLockModel := models.NewTAssetLockModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetLockModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.UnlockAmount(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.AssetUnlockFailed)
		}

		ok, err = assetLockModel.UpdateUnlock(ctx, lock.LockNo, amount, ts)
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

		flow := buildAssetFlowRecord(l.svcCtx, ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, "manual_sub", "system", "manual_sub", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNLOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("AdminUnlockAsset transaction failed, tenantId=%d lockNo=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.LockNo, in.Amount, in.BizNo, err)
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), Data: &asset.AdminChangeAssetData{BizNo: in.BizNo, Asset: toUserAssetProto(after)}}, nil
}
