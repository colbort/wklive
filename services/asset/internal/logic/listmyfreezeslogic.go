package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyFreezesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyFreezesLogic {
	return &ListMyFreezesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的冻结明细
func (l *ListMyFreezesLogic) ListMyFreezes(in *asset.ListMyFreezesReq) (*asset.ListMyFreezesResp, error) {
	items, total, err := l.svcCtx.AssetFreezeModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "", "", int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit
	resp := &asset.ListMyFreezesResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}
	for _, item := range items {
		resp.Data = append(resp.Data, helpers.ToAssetFreezeProto(item))
	}
	return resp, nil
}
