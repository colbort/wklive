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

type AdminUnlockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnlockAssetLogic {
	return &AdminUnlockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台解锁资产
func (l *AdminUnlockAssetLogic) AdminUnlockAsset(in *asset.AdminUnlockAssetReq) (*asset.AdminChangeAssetResp, error) {
	amount, err := helpers.ParseAmount(in.Amount)
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

	ts := time.Now().UnixMilli()
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

	flow := helpers.BuildAssetFlowRecord(lock.TenantId, lock.UserId, lock.WalletType, lock.Coin, "manual_sub", "system", "manual_sub", 0, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_UNLOCK, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.AdminChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
