package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserAssetDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserAssetDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAssetDetailLogic {
	return &GetUserAssetDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户资产详情
func (l *GetUserAssetDetailLogic) GetUserAssetDetail(in *asset.GetUserAssetDetailReq) (*asset.GetUserAssetDetailResp, error) {
	item, err := l.svcCtx.UserAssetModel.FindOneByTenantIdUserIdWalletTypeCoin(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &asset.GetUserAssetDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.AssetNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	return &asset.GetUserAssetDetailResp{Base: helper.OkResp(), Data: toUserAssetProto(item)}, nil
}
