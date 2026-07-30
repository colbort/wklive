package assetlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetBalanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssetBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetBalanceLogic {
	return &GetAssetBalanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 内部查询指定钱包余额，供风险引擎读取统一资产账本。
func (l *GetAssetBalanceLogic) GetAssetBalance(in *asset.GetUserAssetDetailReq) (*asset.GetUserAssetDetailResp, error) {
	if in.TenantId <= 0 || in.UserId <= 0 ||
		in.WalletType == common.WalletType_WALLET_TYPE_UNKNOWN ||
		strings.TrimSpace(in.Coin) == "" {
		return &asset.GetUserAssetDetailResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	item, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(
		l.ctx, in.TenantId, in.UserId, int64(in.WalletType), strings.ToUpper(strings.TrimSpace(in.Coin)),
	)
	if errors.Is(err, models.ErrNotFound) {
		return &asset.GetUserAssetDetailResp{Base: helper.ErrResp(i18n.AssetNotFound, i18n.Translate(i18n.AssetNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset.GetUserAssetDetailResp{Base: helper.OkResp(), Data: helpers.ToUserAssetProto(item)}, nil
}
