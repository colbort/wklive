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
	freezes, _, err := l.svcCtx.AssetFreezeModel.FindPageByFilter(l.ctx, in.TenantId, in.UserId, int64(in.WalletType), in.Coin, "", "", int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	resp := &asset.ListMyFreezesResp{Base: helper.OkResp()}
	for _, item := range freezes {
		resp.Data = append(resp.Data, helpers.ToAssetFreezeProto(item))
	}
	return resp, nil
}
