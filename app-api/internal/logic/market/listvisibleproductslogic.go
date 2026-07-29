// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package market

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/market"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVisibleProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleProductsLogic {
	return &ListVisibleProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVisibleProductsLogic) ListVisibleProducts(req *types.ListVisibleProductsReq) (resp *types.ListVisibleProductsResp, err error) {
	return logicutil.Proxy[types.ListVisibleProductsResp](l.ctx, &market.ListVisibleProductsReq{
		CategoryType: market.CategoryType(req.CategoryType),
		Market:       req.Market,
		Keyword:      req.Keyword,
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
	}, l.svcCtx.MarketCli.ListVisibleProducts)
}
