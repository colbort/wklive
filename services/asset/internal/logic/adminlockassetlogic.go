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

type AdminLockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLockAssetLogic {
	return &AdminLockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台锁仓资产
func (l *AdminLockAssetLogic) AdminLockAsset(in *asset.AdminLockAssetReq) (*asset.AdminChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
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
			return fmt.Errorf("insufficient available balance")
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		lock = buildAssetLockRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "system", "manual_add", in.BizNo, in.Remark, amount, 0, 0, ts)
		if _, err := assetLockModel.Insert(ctx, lock); err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_add", "system", "manual_add", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_LOCK, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
