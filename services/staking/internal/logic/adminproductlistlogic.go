package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductListLogic {
	return &AdminProductListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品列表
func (l *AdminProductListLogic) AdminProductList(in *staking.AdminProductListReq) (*staking.AdminProductListResp, error) {
	page := in.GetPage()
	cursor, limit := int64(0), int64(10)
	if page != nil {
		cursor, limit = page.Cursor, page.Limit
	}
	items, total, err := l.svcCtx.StakeProductModel.FindPage(
		l.ctx, in.TenantId, cursor, limit,
		in.ProductNo, in.ProductName, in.CoinSymbol,
		int64(in.ProductType), int64(in.Status),
	)
	if err != nil {
		return nil, err
	}

	resp := &staking.AdminProductListResp{Page: helper.OkResp()}
	if len(items) == 0 {
		resp.Page = pageutil.Base(cursor, limit, 0, total, 0)
		return resp, nil
	}

	resp.Data = make([]*staking.StakeProduct, 0, len(items))
	for _, item := range items {
		resp.Data = append(resp.Data, productToProto(item))
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, int64(items[len(items)-1].Id))
	return resp, nil
}
