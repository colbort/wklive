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

type AdminFreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminFreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFreezeAssetLogic {
	return &AdminFreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台冻结资产
func (l *AdminFreezeAssetLogic) AdminFreezeAsset(in *asset.AdminFreezeAssetReq) (*asset.AdminChangeAssetResp, error) {
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

	ts := time.Now().UnixMilli()
	ok, err := l.svcCtx.UserAssetModel.FreezeAmount(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, ts)
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

	freeze := helpers.BuildAssetFreezeRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "system", "manual_add", in.BizNo, in.Remark, amount, 0, ts)
	if _, err := l.svcCtx.AssetFreezeModel.Insert(l.ctx, freeze); err != nil {
		return nil, err
	}

	flow := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "manual_add", "system", "manual_add", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
