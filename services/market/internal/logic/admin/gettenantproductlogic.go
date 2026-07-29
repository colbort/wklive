package adminlogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

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
func (l *GetTenantProductLogic) GetTenantProduct(in *market.GetTenantProductReq) (*market.GetTenantProductResp, error) {
	item, err := l.svcCtx.MarketTenantProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.TenantId != in.TenantId {
		return &market.GetTenantProductResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	product, err := l.svcCtx.MarketProductModel.FindOne(l.ctx, item.ProductId)
	if err != nil {
		return nil, err
	}

	return &market.GetTenantProductResp{
		Base: helper.OkResp(),
		Data: helpers.ToTenantProductProto(item, product),
	}, nil
}
