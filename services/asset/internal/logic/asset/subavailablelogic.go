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

type SubAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubAvailableLogic {
	return &SubAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减可用余额
func (l *SubAvailableLogic) SubAvailable(in *asset.SubAvailableReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		return nil, err
	}
	if !amount.IsPositive() {
		return nil, i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotentModel := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		if in.BizNo != "" {
			done, err := helpers.PrepareAssetIdempotent(ctx, idempotentModel, in.TenantId, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, ts)
			if err != nil {
				return err
			}
			if done {
				after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
				return err
			}
		}

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.SubAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, ts)
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

		flow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetSceneType(in.SceneType), helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_SUB, amount, before, after, in.Remark, ts)
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
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}}, nil
}
