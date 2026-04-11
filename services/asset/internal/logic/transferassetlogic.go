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
	amount, err := conv.ParseFloatField(in.Amount)
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

	if ok, err := l.svcCtx.UserAssetModel.SubAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, amount, utils.NowMillis()); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("insufficient available balance")
	}

	if _, err := l.svcCtx.UserAssetModel.AddAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, amount, 0, utils.NowMillis()); err != nil {
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

	sceneType := assetSceneType(in.SceneType)
	bizType := assetBizType(in.BizType)
	flowOut := buildAssetFlowRecord(l.svcCtx, l.ctx, in.TenantId, in.UserId, int64(in.FromWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_OUT, amount, beforeFrom, afterFrom, in.Remark, utils.NowMillis())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flowOut); err != nil {
		return nil, err
	}

	flowIn := buildAssetFlowRecord(l.svcCtx, l.ctx, in.TenantId, in.UserId, int64(in.ToWalletType), in.Coin, sceneType, bizType, sceneType, in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_TRANSFER_IN, amount, beforeTo, afterTo, in.Remark, utils.NowMillis())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flowIn); err != nil {
		return nil, err
	}

	return &asset.TransferAssetResp{Base: helper.OkResp(), FromAsset: toUserAssetProto(afterFrom), ToAsset: toUserAssetProto(afterTo)}, nil
}
