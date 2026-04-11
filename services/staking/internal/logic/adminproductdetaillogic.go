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

type AdminProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductDetailLogic {
	return &AdminProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品详情
func (l *AdminProductDetailLogic) AdminProductDetail(in *staking.AdminProductDetailReq) (*staking.AdminProductDetailResp, error) {
	item, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AdminProductDetailResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return &staking.AdminProductDetailResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}

	return &staking.AdminProductDetailResp{Page: helper.OkResp(), Data: productToProto(item)}, nil
}
