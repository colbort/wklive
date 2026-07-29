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

type GetTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantCategoryLogic {
	return &GetTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户产品类型详情
func (l *GetTenantCategoryLogic) GetTenantCategory(in *market.GetTenantCategoryReq) (*market.GetTenantCategoryResp, error) {
	item, err := l.svcCtx.MarketTenantCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.TenantId != in.TenantId {
		return &market.GetTenantCategoryResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	category, err := l.svcCtx.MarketCategoryModel.FindOne(l.ctx, item.CategoryId)
	if err != nil {
		return nil, err
	}

	return &market.GetTenantCategoryResp{
		Base: helper.OkResp(),
		Data: helpers.ToTenantCategoryProto(item, category),
	}, nil
}
