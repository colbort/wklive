package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductListLogic {
	return &AppProductListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品列表
func (l *AppProductListLogic) AppProductList(in *staking.AppProductListReq) (*staking.AppProductListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeProductModel.FindPage(
		l.ctx, in.TenantId, cursor, limit,
		"", "", in.CoinSymbol,
		int64(in.ProductType), int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE),
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.AppProductListResp{Base: helper.OkResp()}
	if len(items) == 0 {
		resp.Base = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}
	resp.Data = make([]*staking.StakeProduct, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, productToProto(item))
	}
	resp.Base = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
