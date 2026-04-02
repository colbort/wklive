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

type UpdateProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProductLogic) UpdateProduct(req *types.UpdateProductReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.ItickCli.UpdateProduct(l.ctx, &itick.UpdateProductReq{
		Id:          req.Id,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		BaseCoin:    req.BaseCoin,
		QuoteCoin:   req.QuoteCoin,
		Enabled:     req.Enabled,
		AppVisible:  req.AppVisible,
		Sort:        req.Sort,
		Icon:        req.Icon,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
