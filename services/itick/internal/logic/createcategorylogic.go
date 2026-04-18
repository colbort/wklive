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

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品类型
func (l *CreateCategoryLogic) CreateCategory(in *itick.CreateCategoryReq) (*itick.AdminCommonResp, error) {
	exist, err := l.svcCtx.ItickCategoryModel.FindOneByCategoryType(l.ctx, int64(in.CategoryType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return &itick.AdminCommonResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ResourceAlreadyExists, l.ctx)),
		}, nil
	}

	now := cutils.NowMillis()
	_, err = l.svcCtx.ItickCategoryModel.Insert(l.ctx, &models.TItickCategory{
		CategoryType: int64(in.CategoryType),
		CategoryName: in.CategoryName,
		CategoryCode: categoryTypeCode(in.CategoryType),
		Enabled:      in.Enabled,
		AppVisible:   in.AppVisible,
		Sort:         in.Sort,
		Icon:         in.Icon,
		Remark:       in.Remark,
		CreateTimes:  now,
		UpdateTimes:  now,
	})
	if err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}
