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

type SubAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubAvailableLogic {
	return &SubAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减可用余额
func (l *SubAvailableLogic) SubAvailable(in *asset.SubAvailableReq) (*asset.ChangeAssetResp, error) {
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

	ok, err := l.svcCtx.UserAssetModel.SubAvailableAmount(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, amount, utils.NowMillis())
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

	flow := buildAssetFlowRecord(l.svcCtx, l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, assetSceneType(in.SceneType), assetBizType(in.BizType), assetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_SUB, amount, before, after, in.Remark, utils.NowMillis())
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: toUserAssetProto(after)}, nil
}
