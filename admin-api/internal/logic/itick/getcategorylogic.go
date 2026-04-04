// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCategoryLogic) GetCategory(req *types.GetCategoryReq) (resp *types.GetCategoryResp, err error) {
	result, err := l.svcCtx.ItickCli.GetCategory(l.ctx, &itick.GetCategoryReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetCategoryResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.ItickCategory{
			Id:           result.Data.Id,
			CategoryType: int64(result.Data.CategoryType),
			CategoryCode: result.Data.CategoryCode,
			CategoryName: result.Data.CategoryName,
			Enabled:      result.Data.Enabled,
			AppVisible:   result.Data.AppVisible,
			Sort:         result.Data.Sort,
			Icon:         result.Data.Icon,
			Remark:       result.Data.Remark,
			CreateTimes:  result.Data.CreateTimes,
			UpdateTimes:  result.Data.UpdateTimes,
		},
	}, nil
}
