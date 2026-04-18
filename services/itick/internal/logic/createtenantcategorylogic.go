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

type CreateTenantCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantCategoryLogic {
	return &CreateTenantCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户产品类型
func (l *CreateTenantCategoryLogic) CreateTenantCategory(in *itick.CreateTenantCategoryReq) (*itick.AdminCommonResp, error) {
	category, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, in.CategoryId)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	exist, err := l.svcCtx.ItickTenantCategoryModel.FindOneByTenantIdCategoryId(l.ctx, in.TenantId, in.CategoryId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.ItickTenantCategoryModel.Insert(l.ctx, &models.TItickTenantCategory{
		TenantId:    in.TenantId,
		CategoryId:  in.CategoryId,
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
