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

type CreateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProductLogic) CreateProduct(req *types.CreateProductReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.ItickCli.CreateProduct(l.ctx, &itick.CreateProductReq{
		CategoryType: itick.CategoryType(req.CategoryType),
		Market:       req.Market,
		Symbol:       req.Symbol,
		Code:         req.Code,
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		BaseCoin:     req.BaseCoin,
		QuoteCoin:    req.QuoteCoin,
		Enabled:      req.Enabled,
		AppVisible:   req.AppVisible,
		Sort:         req.Sort,
		Icon:         req.Icon,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
