package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayProductLogic {
	return &GetPayProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品详情
func (l *GetPayProductLogic) GetPayProduct(in *payment.GetPayProductReq) (*payment.GetPayProductResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetPayProductResp{}, nil
}
