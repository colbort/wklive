package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, tenantId, userId, int64(in.WalletType), in.Coin)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.GetMyAssetResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.AssetNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	return &asset.GetMyAssetResp{Base: helper.OkResp(), Asset: toUserAssetProto(item)}, nil
}
