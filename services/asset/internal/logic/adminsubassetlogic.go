package logic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
	amount, err := helpers.ParseAmount(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	before, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil {
		return nil, err
	}

	ok, err := l.svcCtx.UserAssetModel.SubAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("insufficient available balance")
	}

	after, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil {
		return nil, err
	}

	flow := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_sub", "system", "manual_sub", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_SUB, amount, before, after, in.Remark, time.Now().UnixMilli())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
