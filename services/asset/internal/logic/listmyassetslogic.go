package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyAssetsLogic {
	return &ListMyAssetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的资产列表
func (l *ListMyAssetsLogic) ListMyAssets(in *asset.ListMyAssetsReq) (*asset.ListMyAssetsResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	list, err := l.svcCtx.UserAssetModel.FindAll(l.ctx, tenantId, userId, int64(in.WalletType), in.Coin, 0)
	if err != nil {
		return nil, err
	}

	resp := &asset.ListMyAssetsResp{Base: helper.OkResp()}
	for _, item := range list {
		resp.Data = append(resp.Data, toUserAssetProto(item))
	}
	return resp, nil
}
