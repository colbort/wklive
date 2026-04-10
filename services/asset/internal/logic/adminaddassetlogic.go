package logic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
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

	ts := time.Now().UnixMilli()
	before, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	if _, err := l.svcCtx.UserAssetModel.AddAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, 0, ts); err != nil {
		return nil, err
	}

	after, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil {
		return nil, err
	}

	flow := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_add", "system", "manual_add", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
