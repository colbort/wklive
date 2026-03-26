package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayPlatformsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayPlatformsLogic {
	return &ListPayPlatformsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 平台列表
func (l *ListPayPlatformsLogic) ListPayPlatforms(in *payment.ListPayPlatformsReq) (*payment.ListPayPlatformsResp, error) {
	// todo: add your logic here and delete this line

	return &payment.ListPayPlatformsResp{}, nil
}
