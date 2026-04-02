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

type GetProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductLogic) GetProduct(req *types.GetProductReq) (resp *types.GetProductResp, err error) {
	result, err := l.svcCtx.ItickCli.GetProduct(l.ctx, &itick.GetProductReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetProductResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.ItickProduct{
			Id:           result.Data.Id,
			CategoryType: int64(result.Data.CategoryType),
			Market:       result.Data.Market,
			Symbol:       result.Data.Symbol,
			Code:         result.Data.Code,
			Name:         result.Data.Name,
			DisplayName:  result.Data.DisplayName,
			BaseCoin:     result.Data.BaseCoin,
			QuoteCoin:    result.Data.QuoteCoin,
			Enabled:      result.Data.Enabled,
			AppVisible:   result.Data.AppVisible,
			Sort:         result.Data.Sort,
			Icon:         result.Data.Icon,
			Remark:       result.Data.Remark,
			CreateTime:   result.Data.CreateTime,
			UpdateTime:   result.Data.UpdateTime,
		},
	}, nil
}
