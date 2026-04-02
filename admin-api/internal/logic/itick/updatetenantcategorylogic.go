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

type UpdateTenantCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantCategoryLogic {
	return &UpdateTenantCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantCategoryLogic) UpdateTenantCategory(req *types.UpdateTenantCategoryReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.ItickCli.UpdateTenantCategory(l.ctx, &itick.UpdateTenantCategoryReq{
		Id:         req.Id,
		TenantId:   req.TenantId,
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
