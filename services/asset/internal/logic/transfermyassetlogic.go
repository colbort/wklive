package logic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/conv"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferMyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTransferMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferMyAssetLogic {
	return &TransferMyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 我的账户划转
func (l *TransferMyAssetLogic) TransferMyAsset(in *asset.TransferMyAssetReq) (*asset.TransferMyAssetResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}

	coin := strings.ToUpper(strings.TrimSpace(in.Coin))
	if coin == "" {
		return nil, fmt.Errorf("coin is required")
	}
	if in.FromWalletType == asset.WalletType_WALLET_TYPE_UNKNOWN || in.ToWalletType == asset.WalletType_WALLET_TYPE_UNKNOWN {
		return nil, fmt.Errorf("wallet type is required")
	}
	if in.FromWalletType == in.ToWalletType {
		return nil, fmt.Errorf("from wallet type and to wallet type must be different")
	}
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	bizNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "TRANSFER", "")
	if err != nil {
		return nil, err
	}
	result, err := NewTransferAssetLogic(l.ctx, l.svcCtx).TransferAsset(&asset.TransferAssetReq{
		TenantId:       tenantId,
		UserId:         userId,
		FromWalletType: in.FromWalletType,
		ToWalletType:   in.ToWalletType,
		Coin:           coin,
		Amount:         in.Amount,
		BizType:        asset.BizType_BIZ_TYPE_TRANSFER,
		SceneType:      asset.SceneType_SCENE_TYPE_TRANSFER,
		BizNo:          bizNo,
		Remark:         in.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &asset.TransferMyAssetResp{
		Base:      result.GetBase(),
		FromAsset: result.GetFromAsset(),
		ToAsset:   result.GetToAsset(),
	}, nil
}
