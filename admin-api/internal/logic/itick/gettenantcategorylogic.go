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

type GetTenantCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantCategoryLogic {
	return &GetTenantCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantCategoryLogic) GetTenantCategory(req *types.GetTenantCategoryReq) (resp *types.GetTenantCategoryResp, err error) {
	result, err := l.svcCtx.ItickCli.GetTenantCategory(l.ctx, &itick.GetTenantCategoryReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetTenantCategoryResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.ItickTenantCategory{
			Id:           result.Data.Id,
			TenantId:     result.Data.TenantId,
			CategoryId:   result.Data.CategoryId,
			Enabled:      result.Data.Enabled,
			AppVisible:   result.Data.AppVisible,
			Sort:         result.Data.Sort,
			Remark:       result.Data.Remark,
			CreateTime:   result.Data.CreateTime,
			UpdateTime:   result.Data.UpdateTime,
			CategoryType: int64(result.Data.CategoryType),
			CategoryName: result.Data.CategoryName,
			Icon:         result.Data.Icon,
		},
	}, nil
}
