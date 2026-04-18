package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantProductLogic {
	return &GetTenantProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户产品详情
func (l *GetTenantProductLogic) GetTenantProduct(in *itick.GetTenantProductReq) (*itick.GetTenantProductResp, error) {
	item, err := l.svcCtx.ItickTenantProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.TenantId != in.TenantId {
		return &itick.GetTenantProductResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	product, err := l.svcCtx.ItickProductModel.FindOne(l.ctx, item.ProductId)
	if err != nil {
		return nil, err
	}

	return &itick.GetTenantProductResp{
		Base: helper.OkResp(),
		Data: toTenantProductProto(item, product),
	}, nil
}
