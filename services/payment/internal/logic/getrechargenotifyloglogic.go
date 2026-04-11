package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

type GetRechargeNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRechargeNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRechargeNotifyLogLogic {
	return &GetRechargeNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 充值回调日志
func (l *GetRechargeNotifyLogLogic) GetRechargeNotifyLog(in *payment.GetRechargeNotifyLogReq) (*payment.GetRechargeNotifyLogResp, error) {
	var (
		errLogic = "GetRechargeNotifyLog"
	)

	notifyLog, err := l.svcCtx.RechargeNotifyLogModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if notifyLog == nil {
		return &payment.GetRechargeNotifyLogResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyLogNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetRechargeNotifyLogResp{
		Base: helper.OkResp(),
		Data: &payment.PayNotifyLog{
			Id:            notifyLog.Id,
			OrderId:       notifyLog.OrderId.Int64,
			OrderNo:       notifyLog.OrderNo.String,
			PlatformId:    notifyLog.PlatformId,
			ChannelId:     notifyLog.ChannelId.Int64,
			NotifyStatus:  payment.NotifyProcessStatus(notifyLog.NotifyStatus),
			NotifyBody:    notifyLog.NotifyBody.String,
			SignResult:    payment.SignResult(notifyLog.SignResult),
			ProcessResult: notifyLog.ProcessResult.String,
			ErrorMessage:  notifyLog.ErrorMessage.String,
			NotifyTime:    notifyLog.NotifyTime.Int64,
			CreateTimes:   notifyLog.CreateTimes,
		},
	}, nil
}
