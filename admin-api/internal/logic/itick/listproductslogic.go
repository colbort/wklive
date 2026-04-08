// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProductsLogic) ListProducts(req *types.ListProductsReq) (resp *types.ListProductsResp, err error) {
	result, err := l.svcCtx.ItickCli.ListProducts(l.ctx, &itick.ListProductsReq{
		Page: &common.PageReq{
			Cursor: req.PageReq.Cursor,
			Limit:  req.PageReq.Limit,
		},
		CategoryType: itick.CategoryType(req.CategoryType),
		CategoryName: req.CategoryName,
		Market:       req.Market,
		Keyword:      req.Keyword,
		Enabled:      req.Enabled,
		AppVisible:   req.AppVisible,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.ItickProduct, 0)
	for _, item := range result.Data {
		data = append(data, types.ItickProduct{
			Id:           item.Id,
			CategoryType: int64(item.CategoryType),
			CategoryName: item.CategoryName,
			CategoryCode: item.CategoryCode,
			Market:       item.Market,
			Symbol:       item.Symbol,
			Code:         item.Code,
			Name:         item.Name,
			DisplayName:  item.DisplayName,
			BaseCoin:     item.BaseCoin,
			QuoteCoin:    item.QuoteCoin,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Icon:         item.Icon,
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		})
	}
	return &types.ListProductsResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}, nil
}
