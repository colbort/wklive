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

type FreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeAssetLogic {
	return &FreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 冻结余额
func (l *FreezeAssetLogic) FreezeAsset(in *asset.FreezeAssetReq) (*asset.FreezeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	ts := utils.NowMillis()
	var (
		after  *models.TUserAsset
		freeze *models.TAssetFreeze
	)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFreezeModel := models.NewTAssetFreezeModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFreezeModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.FreezeAmount(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, ts)
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

		freeze = buildAssetFreezeRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.ExpireTime, ts)
		if _, err := assetFreezeModel.Insert(ctx, freeze); err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &asset.FreezeAssetResp{Base: helper.OkResp(), FreezeNo: freeze.FreezeNo, Asset: toUserAssetProto(after)}, nil
}
