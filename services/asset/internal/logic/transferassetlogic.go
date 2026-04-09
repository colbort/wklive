package logic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferAssetLogic {
	return &TransferAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 钱包划转
func (l *TransferAssetLogic) TransferAsset(in *asset.TransferAssetReq) (*asset.TransferAssetResp, error) {
	amount, err := helpers.ParseAmount(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	beforeFrom, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin)
	if err != nil {
		return nil, err
	}

	if ok, err := l.svcCtx.UserAssetModel.SubAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, amount, time.Now().UnixMilli()); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("insufficient available balance")
	}

	if _, err := l.svcCtx.UserAssetModel.AddAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, amount, 0, time.Now().UnixMilli()); err != nil {
		return nil, err
	}

	beforeTo, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	afterFrom, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin)
	if err != nil {
		return nil, err
	}
	afterTo, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin)
	if err != nil {
		return nil, err
	}

	sceneType := helpers.AssetSceneType(in.SceneType)
	bizType := helpers.AssetBizType(in.BizType)
	flowOut := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_OUT, amount, beforeFrom, afterFrom, in.Remark, time.Now().UnixMilli())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flowOut); err != nil {
		return nil, err
	}

	flowIn := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_IN, amount, beforeTo, afterTo, in.Remark, time.Now().UnixMilli())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flowIn); err != nil {
		return nil, err
	}

	return &asset.TransferAssetResp{Base: helper.OkResp(), FromAsset: helpers.ToUserAssetProto(afterFrom), ToAsset: helpers.ToUserAssetProto(afterTo)}, nil
}
