package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新产品类型仅允许更新名称、状态、排序、图标和备注，产品类型不允许修改
func (l *UpdateCategoryLogic) UpdateCategory(in *itick.UpdateCategoryReq) (*itick.AdminCommonResp, error) {
	item, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.CategoryNotFound, l.ctx)),
		}, nil
	}

	item.CategoryName = in.CategoryName
	item.Enabled = in.Enabled
	item.AppVisible = in.AppVisible
	item.Sort = in.Sort
	item.Icon = in.Icon
	item.Remark = in.Remark
	item.UpdateTimes = cutils.NowMillis()

	if err := l.svcCtx.ItickCategoryModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
