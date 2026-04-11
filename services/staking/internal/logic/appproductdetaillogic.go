package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductDetailLogic {
	return &AppProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品详情
func (l *AppProductDetailLogic) AppProductDetail(in *staking.AppProductDetailReq) (*staking.AppProductDetailResp, error) {
	item, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AppProductDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId || item.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) {
		return &staking.AppProductDetailResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}

	return &staking.AppProductDetailResp{Base: helper.OkResp(), Data: productToProto(item)}, nil
}
