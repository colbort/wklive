package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/asset"
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
	items, total, err := l.svcCtx.AssetFreezeModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "", "", int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	resp := &asset.ListMyFreezesResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID)}
	for _, item := range items {
		resp.Data = append(resp.Data, toAssetFreezeProto(item))
	}
	return resp, nil
}
