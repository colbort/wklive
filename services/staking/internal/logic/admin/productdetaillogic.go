package adminlogic

import (
	"context"
	"errors"
	"wklive/services/staking/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品详情
func (l *ProductDetailLogic) ProductDetail(in *staking.ProductDetailReq) (*staking.ProductDetailResp, error) {
	tenantId, base, err := helpers.AdminTenantReadScopeResp(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if base != nil {
		return &staking.ProductDetailResp{Page: base}, nil
	}
	item, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.ProductDetailResp{Page: helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if tenantId > 0 && item.TenantId != tenantId {
		return &staking.ProductDetailResp{Page: helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}

	return &staking.ProductDetailResp{Page: helper.OkResp(), Data: helpers.ProductToProto(item)}, nil
}
