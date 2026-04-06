// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

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
	result, err := l.svcCtx.ItickCli.ListVisibleProducts(l.ctx, &itick.ListVisibleProductsReq{
		CategoryType: itick.CategoryType(req.CategoryType),
		Market:       req.Market,
		Keyword:      req.Keyword,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.ListVisibleProductsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: make([]types.ItickTenantProduct, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		resp.Data = append(resp.Data, types.ItickTenantProduct{
			Id:           item.Id,
			TenantId:     item.TenantId,
			ProductId:    item.ProductId,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
			CategoryType: int64(item.CategoryType),
			CategoryName: item.CategoryName,
			Market:       item.Market,
			Symbol:       item.Symbol,
			Code:         item.Code,
			Name:         item.Name,
			DisplayName:  item.DisplayName,
			BaseCoin:     item.BaseCoin,
			QuoteCoin:    item.QuoteCoin,
			Icon:         item.Icon,
		})
	}

	return
}
