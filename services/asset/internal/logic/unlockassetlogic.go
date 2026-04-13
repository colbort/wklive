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

type UnlockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockAssetLogic {
	return &UnlockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解锁
func (l *UnlockAssetLogic) UnlockAsset(in *asset.UnlockAssetReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	lock, err := l.svcCtx.AssetLockModel.FindOneByLockNo(l.ctx, in.LockNo)
	if err != nil {
		return nil, err
	}
	if lock.TenantId != in.TenantId {
		return nil, fmt.Errorf("tenant mismatch for lock record")
	}
	if amount > lock.RemainAmount {
		return nil, fmt.Errorf("unlock amount exceeds locked amount")
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
			return fmt.Errorf("unlock failed")
		}

		ok, err = assetLockModel.UpdateUnlock(ctx, lock.LockNo, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("lock record update failed")
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
		if err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNLOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
