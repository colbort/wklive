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

type AdminUnfreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUnfreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnfreezeAssetLogic {
	return &AdminUnfreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台解冻资产
func (l *AdminUnfreezeAssetLogic) AdminUnfreezeAsset(in *asset.AdminUnfreezeAssetReq) (*asset.AdminChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	freeze, err := l.svcCtx.AssetFreezeModel.FindOneByFreezeNo(l.ctx, in.FreezeNo)
	if err != nil {
		return nil, err
	}
	if freeze.TenantId != in.TenantId {
		return nil, fmt.Errorf("tenant mismatch for freeze record")
	}
	if amount > freeze.RemainAmount {
		return nil, fmt.Errorf("unfreeze amount exceeds remaining frozen amount")
	}

	ts := utils.NowMillis()
	var after *models.TUserAsset
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userAssetModel := models.NewTUserAssetModel(conn, l.svcCtx.Config.CacheRedis).(models.UserAssetModel)
		assetFreezeModel := models.NewTAssetFreezeModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFreezeModel)
		assetFlowModel := models.NewTAssetFlowModel(conn, l.svcCtx.Config.CacheRedis).(models.AssetFlowModel)

		before, err := userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
		if err != nil {
			return err
		}

		ok, err := userAssetModel.UnfreezeAmount(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("unfreeze failed")
		}

		ok, err = assetFreezeModel.UpdateUnfreeze(ctx, freeze.FreezeNo, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("freeze record update failed")
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
		if err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, "manual_sub", "system", "manual_sub", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNFREEZE, amount, before, after, in.Remark, ts)
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
