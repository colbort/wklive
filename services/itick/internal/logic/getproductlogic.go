package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品详情
func (l *GetProductLogic) GetProduct(in *itick.GetProductReq) (*itick.GetProductResp, error) {
	result, err := l.svcCtx.ItickProductModel.FindOne(l.ctx, in.Id)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if result == nil {
		return &itick.GetProductResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotFound, l.ctx)),
		}, nil
	}
	return &itick.GetProductResp{
		Base: helper.OkResp(),
		Data: toProductProto(result),
	}, nil
}
