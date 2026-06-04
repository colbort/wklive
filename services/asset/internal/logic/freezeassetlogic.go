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
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("FreezeAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
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
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFreezeModel := models.NewTAssetFreezeModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFreezeModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

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

		freeze = buildAssetFreezeRecord(l.svcCtx, ctx, in.TenantId, in.UserId, walletType, in.Coin, assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.ExpireTime, ts)
		if _, err := assetFreezeModel.Insert(ctx, freeze); err != nil {
			l.Errorf("FreezeAsset insert freeze record failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s freezeNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, freeze.FreezeNo, err)
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, walletType, in.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			l.Errorf("FreezeAsset insert asset flow failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s freezeNo=%s err=%v",
				in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, freeze.FreezeNo, err)
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("FreezeAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizType=%d sceneType=%d bizId=%d bizNo=%s err=%v",
			in.TenantId, in.UserId, walletType, in.Coin, in.Amount, in.BizType, in.SceneType, in.BizId, in.BizNo, err)
		return nil, err
	}

	return &asset.FreezeAssetResp{Base: helper.OkResp(), Data: &asset.FreezeAssetData{FreezeNo: freeze.FreezeNo, Asset: toUserAssetProto(after)}}, nil
}
