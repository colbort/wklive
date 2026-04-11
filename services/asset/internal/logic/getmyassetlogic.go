package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"
)

type GetMyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAssetLogic {
	return &GetMyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的单个币种资产
func (l *GetMyAssetLogic) GetMyAsset(in *asset.GetMyAssetReq) (*asset.GetMyAssetResp, error) {
	item, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.GetMyAssetResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.AssetNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	return &asset.GetMyAssetResp{Base: helper.OkResp(), Asset: helpers.ToUserAssetProto(item)}, nil
}
