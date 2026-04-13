package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantPayChannelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantPayChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayChannelLogic {
	return &GetTenantPayChannelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取租户支付通道详情
func (l *GetTenantPayChannelLogic) GetTenantPayChannel(in *payment.GetTenantPayChannelReq) (*payment.GetTenantPayChannelResp, error) {
	var (
		errLogic = "GetTenantPayChannel"
	)

	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if channel == nil {
		return &payment.GetTenantPayChannelResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.ChannelNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetTenantPayChannelResp{
		Base: helper.OkResp(),
		Data: toTenantPayChannelProto(channel),
	}, nil
}
