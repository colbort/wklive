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

type AddAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAvailableLogic {
	return &AddAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增加可用余额
func (l *AddAvailableLogic) AddAvailable(in *asset.AddAvailableReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset

	changeType := assetSceneType(in.SceneType)
	if changeType == "" {
		changeType = assetBizType(in.BizType)
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)
		idempotentModel := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetIdempotentModel)

		if in.BizNo != "" {
			done, err := prepareAssetIdempotent(ctx, idempotentModel, in.TenantId, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizNo, in.Remark, ts)
			if err != nil {
				return err
			}
			if done {
				after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
				return err
			}
		}

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil && err != models.ErrNotFound {
			return err
		}

		if before == nil {
			_, err = userAssetModel.Insert(ctx, &models.TUserAsset{
				TenantId:        in.TenantId,
				UserId:          in.UserId,
				WalletType:      int64(in.WalletType),
				Coin:            in.Coin,
				TotalAmount:     amount,
				AvailableAmount: amount,
				FrozenAmount:    0,
				LockedAmount:    0,
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
			if _, err := userAssetModel.AddAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, ts); err != nil {
				return err
			}
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, changeType, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		if in.BizNo != "" {
			if err := completeAssetIdempotent(ctx, idempotentModel, in.TenantId, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizNo, ts); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
