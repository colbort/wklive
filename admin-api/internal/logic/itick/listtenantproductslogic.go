// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantProductsLogic {
	return &ListTenantProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantProductsLogic) ListTenantProducts(req *types.ListTenantProductsReq) (resp *types.ListTenantProductsResp, err error) {
	result, err := l.svcCtx.ItickCli.ListTenantProducts(l.ctx, &itick.ListTenantProductsReq{
		Page: &itick.PageReq{
			Cursor: req.Page.Cursor,
			Limit:  req.Page.Limit,
		},
		TenantId:      req.TenantId,
		CategoryType:  itick.CategoryType(req.CategoryType),
		Market:        req.Market,
		Keyword:       req.Keyword,
		Status:        req.Status,
		VisibleStatus: req.VisibleStatus,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.ItickTenantProduct, 0)
	for _, item := range result.Data {
		data = append(data, types.ItickTenantProduct{
			Id:           item.Id,
			TenantId:     item.TenantId,
			ProductId:    item.ProductId,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Remark:       item.Remark,
			CreateTime:   item.CreateTime,
			UpdateTime:   item.UpdateTime,
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

	return &types.ListTenantProductsResp{
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
