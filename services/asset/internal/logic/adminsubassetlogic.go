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

type AdminSubAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminSubAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSubAssetLogic {
	return &AdminSubAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台人工减币
func (l *AdminSubAssetLogic) AdminSubAsset(in *asset.AdminSubAssetReq) (*asset.AdminChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("AdminSubAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("AdminSubAsset validate amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizNo, err)
		return nil, err
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

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

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_sub", "system", "manual_sub", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_SUB, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("AdminSubAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizNo, err)
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), Data: &asset.AdminChangeAssetData{BizNo: in.BizNo, Asset: toUserAssetProto(after)}}, nil
}
