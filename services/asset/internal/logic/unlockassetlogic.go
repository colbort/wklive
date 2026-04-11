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

type UnlockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockAssetLogic {
	return &UnlockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解锁
func (l *UnlockAssetLogic) UnlockAsset(in *asset.UnlockAssetReq) (*asset.ChangeAssetResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	lock, err := l.svcCtx.AssetLockModel.FindOneByLockNo(l.ctx, in.LockNo)
	if err != nil {
		return nil, err
	}
	if lock.TenantId != in.TenantId {
		return nil, fmt.Errorf("tenant mismatch for lock record")
	}
	if amount > lock.RemainAmount {
		return nil, fmt.Errorf("unlock amount exceeds locked amount")
	}

	before, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
	if err != nil {
		return nil, err
	}

	ts := utils.NowMillis()
	ok, err := l.svcCtx.UserAssetModel.UnlockAmount(l.ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, amount, ts)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unlock failed")
	}

	ok, err = l.svcCtx.AssetLockModel.UpdateUnlock(l.ctx, lock.LockNo, amount, ts)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("lock record update failed")
	}

	after, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin)
	if err != nil {
		return nil, err
	}

	flow := buildAssetFlowRecord(l.svcCtx, l.ctx, lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNLOCK, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
