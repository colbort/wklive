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
		return nil, fmt.Errorf("deduct amount exceeds remaining frozen amount")
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

		ok, err := userAssetModel.DeductFromFrozen(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("deduct from frozen failed")
		}

		ok, err = assetFreezeModel.UpdateDeduct(ctx, freeze.FreezeNo, amount, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("freeze record deduct update failed")
		}

		after, err = userAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
		if err != nil {
			return err
		}

		flow := buildAssetFlowRecord(l.svcCtx, ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT, amount, before, after, in.Remark, ts)
		if _, err := assetFlowModel.Insert(ctx, flow); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
