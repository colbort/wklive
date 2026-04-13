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
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
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

		if _, err := userAssetModel.AddAvailableAmount(ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, 0, ts); err != nil {
			return err
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
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
