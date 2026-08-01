package assetlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
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
	walletType := int64(in.WalletType)
	in.BizNo = strings.TrimSpace(in.BizNo)
	if in.BizType == asset.BizType_BIZ_TYPE_OPTION && in.BizNo == "" {
		return nil, errors.New("option asset freeze biz_no is required for idempotency")
	}
	amount, err := conv.ParseDecimalField(in.Amount)
	if err != nil {
		l.Errorf("FreezeAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if !amount.IsPositive() {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("FreezeAsset validate amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var (
		after  *models.TUserAsset
		freeze *models.TAssetFreeze
	)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis)
		assetFreezeModel := models.NewTAssetFreezeModel(conn, l.svcCtx.Config.CacheRedis)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis)
		idempotentModel := models.NewTAssetIdempotentModel(conn, l.svcCtx.Config.CacheRedis)

		if in.BizNo != "" {
			items, total, err := assetFreezeModel.FindPage(ctx, models.AssetFreezePageFilter{
				TenantId: in.TenantId, BizType: helpers.AssetBizType(in.BizType),
				SceneType: helpers.AssetSceneType(in.SceneType), BizNo: in.BizNo,
			}, 0, 2)
			if err != nil {
				return err
			}
			if total > 1 || len(items) > 1 {
				return errors.New("asset freeze business key has duplicate freeze evidence")
			}
			if total == 1 && len(items) == 1 {
				freeze = items[0]
				if !assetFreezeReplayMatches(freeze, in, amount) {
					return errors.New("asset freeze idempotency key was reused with different economic fields")
				}
				done, err := helpers.PrepareAssetIdempotent(
					ctx, idempotentModel, in.TenantId,
					helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType),
					in.BizNo, in.Remark, ts,
				)
				if err != nil {
					return err
				}
				if !done {
					if err := helpers.CompleteAssetIdempotent(
						ctx, idempotentModel, in.TenantId,
						helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType),
						in.BizNo, ts,
					); err != nil {
						return err
					}
				}
				after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(
					ctx, in.TenantId, in.UserId, walletType, in.Coin,
				)
				return err
			}
			done, err := helpers.PrepareAssetIdempotent(
				ctx, idempotentModel, in.TenantId,
				helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType),
				in.BizNo, in.Remark, ts,
			)
			if err != nil {
				return err
			}
			if done {
				return errors.New("idempotent asset freeze has no freeze evidence")
			}
		}

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, walletType, in.Coin)
		if err != nil {
			l.Errorf("FreezeAsset find asset before freeze failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
			return err
		}

		ok, err := userAssetModel.FreezeAmount(ctx, in.TenantId, in.UserId, walletType, in.Coin, amount, ts)
		if err != nil {
			l.Errorf("FreezeAsset freeze amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
			return err
		}
		if !ok {
			err := i18n.StatusError(ctx, i18n.InsufficientAvailableBalance)
			l.Errorf("FreezeAsset freeze amount insufficient balance, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
			return err
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, in.TenantId, in.UserId, walletType, in.Coin)
		if err != nil {
			l.Errorf("FreezeAsset find asset after freeze failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
			return err
		}

		freeze = helpers.BuildAssetFreezeRecord(l.svcCtx, ctx, in.TenantId, in.UserId, walletType, in.Coin, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.ExpireTime, ts)
		if freeze == nil {
			return errors.New("generate asset freeze number failed")
		}
		freeze.BizId = in.BizId
		if _, err := assetFreezeModel.Insert(ctx, freeze); err != nil {
			l.Errorf("FreezeAsset insert freeze record failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s freezeNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, freeze.FreezeNo, err)
			return err
		}

		flow := helpers.BuildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, walletType, in.Coin, helpers.AssetSceneType(in.SceneType), helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			l.Errorf("FreezeAsset insert asset flow failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s freezeNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, freeze.FreezeNo, err)
			return err
		}
		if in.BizNo != "" {
			if err := helpers.CompleteAssetIdempotent(
				ctx, idempotentModel, in.TenantId,
				helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType),
				in.BizNo, ts,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.Errorf("FreezeAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return &asset.FreezeAssetResp{Base: helper.OkResp(), Data: &asset.FreezeAssetData{FreezeNo: freeze.FreezeNo, Asset: helpers.ToUserAssetProto(after)}}, nil
}

func assetFreezeReplayMatches(
	freeze *models.TAssetFreeze,
	in *asset.FreezeAssetReq,
	amount decimal.Decimal,
) bool {
	return freeze != nil && in != nil &&
		freeze.TenantId == in.TenantId &&
		freeze.UserId == in.UserId &&
		freeze.WalletType == int64(in.WalletType) &&
		freeze.Coin == in.Coin &&
		freeze.BizType == helpers.AssetBizType(in.BizType) &&
		freeze.SceneType == helpers.AssetSceneType(in.SceneType) &&
		(freeze.BizId == in.BizId || freeze.BizId == 0) &&
		freeze.BizNo == in.BizNo &&
		freeze.Amount.Equal(amount) &&
		freeze.ExpireTime == in.ExpireTime
}
