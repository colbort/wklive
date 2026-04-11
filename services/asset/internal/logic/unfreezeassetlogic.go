package logic

import (
	"context"
	"fmt"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeAssetLogic {
	return &UnfreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解冻余额
func (l *UnfreezeAssetLogic) UnfreezeAsset(in *asset.UnfreezeAssetReq) (*asset.ChangeAssetResp, error) {
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

	before, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
	if err != nil {
		return nil, err
	}

	ts := utils.NowMillis()
	ok, err := l.svcCtx.UserAssetModel.UnfreezeAmount(l.ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, amount, ts)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unfreeze failed")
	}

	ok, err = l.svcCtx.AssetFreezeModel.UpdateUnfreeze(l.ctx, freeze.FreezeNo, amount, ts)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("freeze record update failed")
	}

	after, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin)
	if err != nil {
		return nil, err
	}

	flow := buildAssetFlowRecord(l.svcCtx, l.ctx, freeze.TenantId, freeze.UserId, freeze.WalletType, freeze.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNFREEZE, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
