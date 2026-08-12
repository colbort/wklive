package adminlogic

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

// 后台锁仓资产
func (l *LockAssetLogic) LockAsset(in *asset.ManualLockAssetReq) (*asset.ManualChangeAssetResp, error) {
	if base, err := helpers.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &asset.ManualChangeAssetResp{Base: base}, nil
	}
	bizNo, err := generateManualAssetBizNo(l.ctx, l.svcCtx, manualLockBizNoPrefix)
	if err != nil {
		return nil, err
	}
	amount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		l.Errorf("AdminLockAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, bizNo, err)
		return nil, err
	}
	if !amount.IsPositive() {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("AdminLockAsset validate amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, bizNo, err)
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

		lock = helpers.BuildAssetLockRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "system", "manual_add", bizNo, in.Remark, amount, 0, 0, ts)
		if _, err := assetLockModel.Insert(ctx, lock); err != nil {
			return err
		}

		flow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_add", "system", "manual_add", 0, bizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("AdminLockAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, bizNo, err)
		return nil, err
	}

	return &asset.ManualChangeAssetResp{Base: helper.OkResp(), Data: &asset.ManualChangeAssetData{BizNo: bizNo, Asset: helpers.ToUserAssetProto(after)}}, nil
}
