package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

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
func (l *GetTenantCategoryLogic) GetTenantCategory(in *itick.GetTenantCategoryReq) (*itick.GetTenantCategoryResp, error) {
	item, err := l.svcCtx.ItickTenantCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil || item.TenantId != in.TenantId {
		return &itick.GetTenantCategoryResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}

	category, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, item.CategoryId)
	if err != nil {
		return nil, err
	}

	return &itick.GetTenantCategoryResp{
		Base: helper.OkResp(),
		Data: toTenantCategoryProto(item, category),
	}, nil
}
