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

type GetTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductLogic {
	return &GetTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantProductLogic) GetTenantProduct(req *types.GetTenantProductReq) (resp *types.GetTenantProductResp, err error) {
	result, err := l.svcCtx.ItickCli.GetTenantProduct(l.ctx, &itick.GetTenantProductReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetTenantProductResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.ItickTenantProduct{
			Id:           result.Data.Id,
			TenantId:     result.Data.TenantId,
			ProductId:    result.Data.ProductId,
			Enabled:      result.Data.Enabled,
			AppVisible:   result.Data.AppVisible,
			Sort:         result.Data.Sort,
			Remark:       result.Data.Remark,
			CreateTimes:   result.Data.CreateTimes,
			UpdateTimes:   result.Data.UpdateTimes,
			CategoryType: int64(result.Data.CategoryType),
			CategoryName: result.Data.CategoryName,
			Market:       result.Data.Market,
			Symbol:       result.Data.Symbol,
			Code:         result.Data.Code,
			Name:         result.Data.Name,
			DisplayName:  result.Data.DisplayName,
			BaseCoin:     result.Data.BaseCoin,
			QuoteCoin:    result.Data.QuoteCoin,
			Icon:         result.Data.Icon,
		},
	}, nil
}
