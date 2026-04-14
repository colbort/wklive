package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayPlatformsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayPlatformsLogic {
	return &GetPayPlatformsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统支持的平台
func (l *GetPayPlatformsLogic) GetPayPlatforms(in *payment.AdminEmpty) (*payment.PayPlatformsResp, error) {
	data := make([]*payment.PayPlatformItem, 0)
	data = append(data, &payment.PayPlatformItem{
		PlatformCode: "dongfang",
		PlatformName: "东方支付",
	})
	data = append(data, &payment.PayPlatformItem{
		PlatformCode: "xifang",
		PlatformName: "西方支付",
	})
	return &payment.PayPlatformsResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
