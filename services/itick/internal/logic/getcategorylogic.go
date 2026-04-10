package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品类型详情
func (l *GetCategoryLogic) GetCategory(in *itick.GetCategoryReq) (*itick.GetCategoryResp, error) {
	result, err := l.svcCtx.ItickCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &itick.GetCategoryResp{
			Base: helper.GetErrResp(1, "数据不存在"),
		}, nil
	}
	return &itick.GetCategoryResp{
		Base: helper.OkResp(),
		Data: &itick.ItickCategory{
			Id:           result.Id,
			CategoryType: itick.CategoryType(result.CategoryType),
			CategoryCode: result.CategoryCode,
			CategoryName: result.CategoryName,
			Enabled:      result.Enabled,
			AppVisible:   result.AppVisible,
			Sort:         result.Sort,
			Icon:         result.Icon,
			Remark:       result.Remark,
			CreateTimes:  result.CreateTimes,
			UpdateTimes:  result.UpdateTimes,
		},
	}, nil
}
