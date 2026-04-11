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

type GetWithdrawNotifyLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWithdrawNotifyLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWithdrawNotifyLogLogic {
	return &GetWithdrawNotifyLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取提现回调日志详情
func (l *GetWithdrawNotifyLogLogic) GetWithdrawNotifyLog(in *payment.GetWithdrawNotifyLogReq) (*payment.GetWithdrawNotifyLogResp, error) {
	var (
		errLogic = "GetWithdrawNotifyLog"
	)

	notifyLog, err := l.svcCtx.WithdrawNotifyLogModel.FindOne(l.ctx, in.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if notifyLog == nil {
		return &payment.GetWithdrawNotifyLogResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.NotifyLogNotFound, l.ctx)),
		}, nil
	}

	return &payment.GetWithdrawNotifyLogResp{
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
