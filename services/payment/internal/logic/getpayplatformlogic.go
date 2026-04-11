package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
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
	var (
		errLogic = "GetPayPlatform"
	)

	platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if platform == nil {
		return &payment.GetPayPlatformResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.PlatformNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetPayPlatformResp{
		Base: helper.OkResp(),
		Data: &payment.PayPlatform{
			Id:           platform.Id,
			PlatformCode: platform.PlatformCode,
			PlatformName: platform.PlatformName,
			PlatformType: payment.PlatformType(platform.PlatformType),
			NotifyUrl:    platform.NotifyUrl.String,
			ReturnUrl:    platform.ReturnUrl.String,
			Icon:         platform.Icon.String,
			Status:       payment.CommonStatus(platform.Status),
			Remark:       platform.Remark.String,
			CreateTimes:  platform.CreateTimes,
			UpdateTimes:  platform.UpdateTimes,
		},
	}, nil
}
