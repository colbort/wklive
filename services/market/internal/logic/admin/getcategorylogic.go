package adminlogic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/logic/helpers"
	"wklive/services/market/internal/svc"

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
func (l *GetCategoryLogic) GetCategory(in *market.GetCategoryReq) (*market.GetCategoryResp, error) {
	result, err := l.svcCtx.MarketCategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &market.GetCategoryResp{
			Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}
	return &market.GetCategoryResp{
		Base: helper.OkResp(),
		Data: helpers.ToCategoryProto(result),
	}, nil
}
