package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
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
	if in.Status != asset.AssetStatus_ASSET_STATUS_UNSPECIFIED {
		status = helpers.AssetStatusFilter(in.Status)
	}

	list, total, err := l.svcCtx.UserAssetModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, status, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(list)) == in.Page.Limit && in.Page.Limit > 0 {
		nextCursor = list[len(list)-1].Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(list)) == in.Page.Limit && in.Page.Limit > 0

	resp := &asset.PageUserAssetsResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}

	for _, item := range list {
		resp.Data = append(resp.Data, helpers.ToUserAssetProto(item))
	}
	return resp, nil
}
