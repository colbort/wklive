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
	freezes, total, err := l.svcCtx.AssetFreezeModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, helpers.AssetBizType(in.BizType), in.BizNo, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	resp := &asset.PageAssetFreezesResp{Base: helper.OkResp()}
	resp.Base.Total = total
	if int64(len(freezes)) == in.Page.Limit && in.Page.Limit > 0 {
		resp.Base.HasNext = true
		resp.Base.NextCursor = freezes[len(freezes)-1].Id
	}

	for _, item := range freezes {
		resp.Data = append(resp.Data, helpers.ToAssetFreezeProto(item))
	}
	return resp, nil
}
