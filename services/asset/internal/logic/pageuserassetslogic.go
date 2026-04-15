package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageUserAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageUserAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageUserAssetsLogic {
	return &PageUserAssetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资产
func (l *PageUserAssetsLogic) PageUserAssets(in *asset.PageUserAssetsReq) (*asset.PageUserAssetsResp, error) {
	status := int64(0)
	if in.Status != asset.AssetStatus_ASSET_STATUS_UNKNOWN {
		status = assetStatusFilter(in.Status)
	}

	list, total, err := l.svcCtx.UserAssetModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, status, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = list[len(list)-1].Id
	}

	resp := &asset.PageUserAssetsResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(list), total, lastID)}

	for _, item := range list {
		resp.Data = append(resp.Data, toUserAssetProto(item))
	}
	return resp, nil
}
