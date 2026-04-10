package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/helpers"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetFreezesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFreezesLogic {
	return &PageAssetFreezesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询冻结明细
func (l *PageAssetFreezesLogic) PageAssetFreezes(in *asset.PageAssetFreezesReq) (*asset.PageAssetFreezesResp, error) {
	freezes, total, err := l.svcCtx.AssetFreezeModel.FindPage(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), in.BizNo, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(freezes)) == in.Page.Limit && in.Page.Limit > 0 {
		nextCursor = freezes[len(freezes)-1].Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(freezes)) == in.Page.Limit && in.Page.Limit > 0

	resp := &asset.PageAssetFreezesResp{Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor)}

	for _, item := range freezes {
		resp.Data = append(resp.Data, helpers.ToAssetFreezeProto(item))
	}
	return resp, nil
}
