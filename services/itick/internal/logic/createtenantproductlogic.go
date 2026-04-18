package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantProductLogic {
	return &CreateTenantProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品
func (l *CreateTenantProductLogic) CreateTenantProduct(in *itick.CreateTenantProductReq) (*itick.AdminCommonResp, error) {
	product, err := l.svcCtx.ItickProductModel.FindOne(l.ctx, in.ProductId)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ProductNotFound, l.ctx)),
		}, nil
	}

	exist, err := l.svcCtx.ItickTenantProductModel.FindOneByTenantIdProductId(l.ctx, in.TenantId, in.ProductId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.ItickTenantProductModel.Insert(l.ctx, &models.TItickTenantProduct{
		TenantId:    in.TenantId,
		ProductId:   in.ProductId,
		Enabled:     in.Enabled,
		AppVisible:  in.AppVisible,
		Sort:        in.Sort,
		Remark:      in.Remark,
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
