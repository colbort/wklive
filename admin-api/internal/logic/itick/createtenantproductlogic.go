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

type CreateTenantProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantProductLogic {
	return &CreateTenantProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantProductLogic) CreateTenantProduct(req *types.CreateTenantProductReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.ItickCli.CreateTenantProduct(l.ctx, &itick.CreateTenantProductReq{
		TenantId:   req.TenantId,
		ProductId:  req.ProductId,
		Enabled:    req.Enabled,
		AppVisible: req.AppVisible,
		Sort:       req.Sort,
		Remark:     req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
