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

type AddAvailableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAvailableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAvailableLogic {
	return &AddAvailableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 增加可用余额
func (l *AddAvailableLogic) AddAvailable(in *asset.AddAvailableReq) (*asset.ChangeAssetResp, error) {
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

	changeType := helpers.AssetSceneType(in.SceneType)
	if changeType == "" {
		changeType = helpers.AssetBizType(in.BizType)
	}

	flow := helpers.BuildAssetFlowRecord(in.TenantId, in.UserId, int64(in.WalletType), in.Coin, changeType, helpers.AssetBizType(in.BizType), helpers.AssetSceneType(in.SceneType), in.BizId, in.BizNo, asset.AssetOpType_ASSET_OP_TYPE_ADD, amount, before, after, in.Remark, ts)
	if _, err := l.svcCtx.AssetFlowModel.Insert(l.ctx, flow); err != nil {
		return nil, err
	}

	return &asset.ChangeAssetResp{Base: helper.OkResp(), BizNo: in.BizNo, Asset: helpers.ToUserAssetProto(after)}, nil
}
