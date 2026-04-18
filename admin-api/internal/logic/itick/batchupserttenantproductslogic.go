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

type BatchUpsertTenantProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchUpsertTenantProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchUpsertTenantProductsLogic {
	return &BatchUpsertTenantProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchUpsertTenantProductsLogic) BatchUpsertTenantProducts(req *types.BatchUpsertTenantProductsReq) (resp *types.RespBase, err error) {
	data := make([]*itick.TenantProductItem, 0, len(req.Data))
	for _, item := range req.Data {
		data = append(data, &itick.TenantProductItem{
			Id:         item.Id,
			ProductId:  item.ProductId,
			Enabled:    item.Enabled,
			AppVisible: item.AppVisible,
			Sort:       item.Sort,
			Remark:     item.Remark,
		})
	}

	result, err := l.svcCtx.ItickCli.BatchUpsertTenantProducts(l.ctx, &itick.BatchUpsertTenantProductsReq{
		TenantId: req.TenantId,
		Data:     data,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
