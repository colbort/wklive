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

type CreateTenantCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantCategoryLogic {
	return &CreateTenantCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantCategoryLogic) CreateTenantCategory(req *types.CreateTenantCategoryReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.ItickCli.CreateTenantCategory(l.ctx, &itick.CreateTenantCategoryReq{
		TenantId:   req.TenantId,
		CategoryId: req.CategoryId,
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
