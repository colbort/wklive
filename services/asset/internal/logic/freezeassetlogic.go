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

	"github.com/zeromicro/go-zero/core/logx"
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
	amount, err := conv.ParseFloatField(in.Amount)
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

	freeze := helpers.BuildAssetFreezeRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizNo, in.Remark, amount, in.ExpireTime, ts)
	if _, err := l.svcCtx.AssetFreezeModel.Insert(l.ctx, freeze); err != nil {
		return nil, err
	}

	flow := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetSceneType(in.SceneType), helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_FREEZE, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.FreezeAssetResp{Base: helper.OkResp(), FreezeNo: freeze.FreezeNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
