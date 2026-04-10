package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyLocksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyLocksLogic {
	return &ListMyLocksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的锁仓明细
func (l *ListMyLocksLogic) ListMyLocks(in *asset.ListMyLocksReq) (*asset.ListMyLocksResp, error) {
	items, total, err := l.svcCtx.AssetLockModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "", "", int64(in.Status), in.Page.Cursor, in.Page.Limit)
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
	resp := &asset.ListMyLocksResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}
	for _, item := range items {
		resp.Data = append(resp.Data, helpers.ToAssetLockProto(item))
	}
	return resp, nil
}
