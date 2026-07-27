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

type DeductFrozenAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductFrozenAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductFrozenAssetLogic {
	return &DeductFrozenAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减冻结余额
func (l *DeductFrozenAssetLogic) DeductFrozenAsset(in *asset.DeductFrozenAssetReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		l.Errorf("DeductFrozenAsset parse amount failed, tenantId=%d freezeNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.FreezeNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if !amount.IsPositive() {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("DeductFrozenAsset validate amount failed, tenantId=%d freezeNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.FreezeNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetFreezeModel := models.NewTAssetFreezeModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotentModel := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		freeze, err := assetFreezeModel.FindOneByFreezeNo(ctx, in.FreezeNo)
		if err != nil {
			return err
		}
		if freeze.TenantId != in.TenantId {
			return i18n.StatusError(ctx, i18n.AssetTenantMismatch)
		}
		if in.BizNo != "" {
			done, err := helpers.PrepareAssetIdempotent(ctx, idempotentModel, in.TenantId, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, ts)
			if err != nil {
				return err
			}
			if done {
				after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
				return err
			}
		}
		if freeze.Status != 1 && freeze.Status != 2 {
			return i18n.StatusError(ctx, i18n.FreezeRecordNotDeductible)
		}
		if amount.GreaterThan(freeze.RemainAmount) {
			return i18n.StatusError(ctx, i18n.DeductAmountExceedsFrozen)
		}

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.DeductFromFrozen(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.DeductFrozenFailed)
		}

		ok, err = assetFreezeModel.UpdateDeduct(ctx, freeze.FreezeNo, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return i18n.StatusError(ctx, i18n.FreezeRecordDeductUpdateFailed)
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
		if err != nil {
			return err
		}

		flow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, helpers.AssetSceneType(in.SceneType), helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT, amount, before, after, in.Remark, ts)
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
		l.Errorf("DeductFrozenAsset transaction failed, tenantId=%d freezeNo=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.FreezeNo, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), Data: &asset.ChangeAssetData{BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}}, nil
}
