package logic

import (
	"context"

	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayPlatformLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayPlatformLogic {
	return &GetPayPlatformLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取平台详情
func (l *GetPayPlatformLogic) GetPayPlatform(in *payment.GetPayPlatformReq) (*payment.GetPayPlatformResp, error) {
	// todo: add your logic here and delete this line

	return &payment.GetPayPlatformResp{}, nil
}
