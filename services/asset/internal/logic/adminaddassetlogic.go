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

type AdminAddAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminAddAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAddAssetLogic {
	return &AdminAddAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台人工加币
func (l *AdminAddAssetLogic) AdminAddAsset(in *asset.AdminAddAssetReq) (*asset.AdminChangeAssetResp, error) {
	if base, err := adminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &asset.AdminChangeAssetResp{Base: base}, nil
	}

	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		l.Errorf("AdminAddAsset parse amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizNo, err)
		return nil, err
	}
	if amount <= 0 {
		err := i18n.StatusError(l.ctx, i18n.AmountMustBePositive)
		l.Errorf("AdminAddAsset validate amount failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
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
				Enabled:         1,
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

		flow := buildAssetFlowRecord(l.svcCtx, ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_add", "system", "manual_add", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("AdminAddAsset transaction failed, tenantId=%d userId=%d walletType=%d coin=%s amount=%s bizNo=%s err=%v",
			in.TenantId, in.UserId, in.WalletType, in.Coin, in.Amount, in.BizNo, err)
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), Data: &asset.AdminChangeAssetData{BizNo: in.BizNo, Asset: toUserAssetProto(after)}}, nil
}
