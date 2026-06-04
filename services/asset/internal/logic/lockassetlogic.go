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
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("LockAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
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
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetLockModel := models.NewTAssetLockModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetLockModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

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

		lock = buildAssetLockRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.StartTime, in.EndTime, ts)
		if _, err := assetLockModel.Insert(ctx, lock); err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("LockAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizNo, err)
		return nil, err
	}

	return &asset.LockAssetResp{Base: helper.OkResp(), Data: &asset.LockAssetData{LockNo: lock.LockNo, Asset: toUserAssetProto(after)}}, nil
}
