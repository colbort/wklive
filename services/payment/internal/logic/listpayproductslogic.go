package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayProductsLogic {
	return &ListPayProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品列表
func (l *ListPayProductsLogic) ListPayProducts(in *payment.ListPayProductsReq) (*payment.ListPayProductsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListPayProductsResp{}, nil
}
